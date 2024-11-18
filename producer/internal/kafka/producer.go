package kafka

import (
	"github.com/IBM/sarama"
	"log"
)

func ConnectProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	return sarama.NewSyncProducer(brokers, config)
}

func PushLedgerToQueue(topic string, message []byte) error {
	//brokers := []string{cfg.Kafka.BrokerURL}
	brokers := []string{"kafka:9092"}
	producer, err := ConnectProducer(brokers)
	if err != nil {
		return err
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Ledger is stored in topic (%s)/partition(%d)/offset(%d)\n",
		topic,
		partition,
		offset,
	)

	return nil
}
