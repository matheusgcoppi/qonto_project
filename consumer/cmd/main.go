package main

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"qonto_project/consumer/internal/config"
	"qonto_project/consumer/internal/database"
	"qonto_project/consumer/internal/kafka"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	db, err := database.ConnectDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Could not run migrations: %v", err)
	}

	topics := []string{"transaction_ledger", "initial_balance"}
	var msgCount int32 = 0

	worker, err := kafka.ConnectConsumer([]string{cfg.Kafka.BrokerURL})
	if err != nil {
		panic(err)
	}

	log.Println("Consumer started")

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	doneCh := make(chan struct{})
	var wg sync.WaitGroup

	go func() {
		<-sigchan
		log.Println("Interrupt detected. Shutting down...")
		close(doneCh) // Broadcast shutdown to all goroutines
	}()

	for _, topic := range topics {
		log.Printf("Processing topic: %s", topic)
		partitions, err := worker.Partitions(topic)
		if err != nil {
			log.Printf("Failed to get partitions for topic %s: %v", topic, err)
			continue
		}
		log.Printf("Found %d partitions for topic: %s", len(partitions), topic)

		for _, partition := range partitions {
			wg.Add(1)
			consumer, err := worker.ConsumePartition(topic, partition, sarama.OffsetOldest)
			if err != nil {
				log.Printf("Failed to start consumer for partition %d: %v", partition, err)
				continue
			}

			go func(topic string, partition int32) {
				log.Printf("Started consuming topic %s, partition %d", topic, partition)
				defer consumer.Close()
				defer wg.Done()
				for {
					select {
					case err := <-consumer.Errors():
						log.Println(err)
					case msg := <-consumer.Messages():
						atomic.AddInt32(&msgCount, 1)
						handleMessage(msg, db)
						log.Printf("Received transaction %d, topic: %s, message: %s\n", msgCount, msg.Topic, msg.Value)
					case <-doneCh:
						log.Printf("Shutting down consumer for topic %s, partition %d", topic, partition)
						return
					}
				}
			}(topic, partition)
		}
	}
	wg.Wait()
	log.Printf("Processed %d messages", atomic.LoadInt32(&msgCount))

	if err := worker.Close(); err != nil {
		log.Fatalf("Failed to close Kafka consumer: %v", err)
	}
}

func handleMessage(msg *sarama.ConsumerMessage, db *gorm.DB) {
	log.Printf("Received message from topic: %s, partition: %d, offset: %d", msg.Topic, msg.Partition, msg.Offset)

	switch msg.Topic {
	case "transaction_ledger":
		handleTransactionLedger(msg.Value, db)
	case "initial_balance":
		handleInitialBalance(msg.Value, db)
	default:
		log.Printf("Unknown topic: %s", msg.Topic)
	}
}

func handleTransactionLedger(message []byte, db *gorm.DB) {
	var transaction struct {
		FromID string `json:"from_id"`
		ToID   string `json:"to_id"`
		Amount int    `json:"amount"`
	}

	var fromAccount struct {
		ID      int
		Balance int
	}

	var toAccount struct {
		ID      int
		Balance int
	}

	if err := json.Unmarshal(message, &transaction); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		query := "SELECT id_account AS ID, balance AS Balance FROM balances WHERE id_account = ?"
		if err := tx.Raw(query, transaction.FromID).Scan(&fromAccount).Error; err != nil {
			log.Printf("Error fetching balance for FromID %d: %v", transaction.FromID, err)
			return err
		}

		if fromAccount.Balance < transaction.Amount {
			log.Printf("Insufficient funds for FromID %d. Balance: %d, Required: %d",
				transaction.FromID, fromAccount.Balance, transaction.Amount)

			// TODO: SEND A MESSAGE BACK USING KAFKA THAT THE USER HAS NO SUFFICIENT FUND
			return fmt.Errorf("insufficient funds for FromID %d", transaction.FromID)
		}

		if err := tx.Raw(query, transaction.ToID).Scan(&toAccount).Error; err != nil {
			log.Printf("Error fetching balance for ToID %d: %v", transaction.ToID, err)
			return err
		}

		updateQuery := "UPDATE balances SET balance = ? WHERE id_account = ?"
		if err := tx.Exec(updateQuery, fromAccount.Balance-transaction.Amount, transaction.FromID).Error; err != nil {
			log.Printf("Error updating balance for FromID %d: %v", transaction.FromID, err)
			return err
		}
		if err := tx.Exec(updateQuery, toAccount.Balance+transaction.Amount, transaction.ToID).Error; err != nil {
			log.Printf("Error updating balance for ToID %d: %v", transaction.ToID, err)
			return err
		}

		log.Printf("Transaction completed: FromID %d -> ToID %d, Amount: %d",
			transaction.FromID, transaction.ToID, transaction.Amount)

		return nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
		return
	}

	log.Println("Transaction processed successfully")
	//TODO: SEND A MESSAGE TO KAFKA THAT JUST NOW CAN CREATE A TRANSACTION IN DB
}

func handleInitialBalance(message []byte, db *gorm.DB) {
	var result struct {
		ID int
	}

	var accountMessage struct {
		ID             int `json:"id"`
		InitialBalance int `json:"initial_balance"`
	}

	if err := json.Unmarshal(message, &accountMessage); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	accountID := strconv.Itoa(accountMessage.ID)

	err := db.Transaction(func(tx *gorm.DB) error {
		query := "SELECT id_account AS ID FROM balances WHERE id_account = ?"
		if err := tx.Raw(query, accountID).Scan(&result).Error; err != nil {
			log.Printf("Error querying balances for Account ID %s: %v", accountID, err)
			return err
		}

		if result.ID != 0 {
			log.Printf("Account ID %s already initialized in balance", accountID)
			return fmt.Errorf("account ID %s already exists", accountID)
		}

		insertQuery := "INSERT INTO balances (id_account, balance) VALUES (?, ?)"
		if err := tx.Exec(insertQuery, accountID, accountMessage.InitialBalance).Error; err != nil {
			log.Printf("Failed to create a balance for Account ID %s: %v", accountID, err)
			return err
		}

		log.Printf("Successfully created balance for Account ID %s with Initial Balance %d",
			accountID, accountMessage.InitialBalance)
		return nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
		return
	}
}
