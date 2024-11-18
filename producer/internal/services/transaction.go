package services

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"log"
	"math"
	"net/http"
	"qonto_project/producer/internal/kafka"
	"qonto_project/producer/internal/models"
)

func CreateTransactionHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var input struct {
		FromID string  `json:"from_id"`
		ToID   string  `json:"to_id"`
		Amount float64 `json:"amount"` // Accept as decimal
	}

	w.Header().Set("Content-Type", "application/json")

	// Decode JSON input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid Input "+err.Error(), http.StatusBadRequest)
		return
	}

	// Convert Amount to micro-units
	amountInMicroUnits := int64(math.Round(input.Amount * 1_000_000_000))

	// Validate FromID and ToID existence
	var fromAccount, toAccount models.Account
	if err := db.Where("id = ?", input.FromID).First(&fromAccount).Error; err != nil {
		http.Error(w, "FromID does not exist in the database", http.StatusBadRequest)
		return
	}

	if err := db.Where("id = ?", input.ToID).First(&toAccount).Error; err != nil {
		http.Error(w, "ToID does not exist in the database", http.StatusBadRequest)
		return
	}

	// Create a Transaction instance
	transaction := models.Transaction{
		FromID: input.FromID,
		ToID:   input.ToID,
		Amount: amountInMicroUnits,
	}

	// Validate the transaction structure
	if err := models.ValidateLedger(&transaction); err != nil {
		http.Error(w, "Validation Failed "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Serialize the transaction to bytes for Kafka
	ledgerInBytes, err := json.Marshal(transaction)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Push to Kafka
	if err := kafka.PushLedgerToQueue("transaction_ledger", ledgerInBytes); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond to the client
	response := map[string]interface{}{
		"msg": fmt.Sprintf("Transaction from %s to %s of amount %.2f created successfully!", input.FromID, input.ToID, input.Amount),
	}

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
