package services

import (
	"encoding/json"
	"gorm.io/gorm"
	"log"
	"net/http"
	"qonto_project/producer/internal/kafka"
	"qonto_project/producer/internal/models"
)

func CreateAccountHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var input struct {
		Name           string  `json:"name"`
		InitialBalance float64 `json:"initial_balance"` // Accept as decimal
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	account := models.Account{Name: input.Name}
	if err := account.SetInitialBalanceFromDecimal(input.InitialBalance); err != nil {
		http.Error(w, "Invalid balance: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := models.ValidateAccount(&account); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := db.Create(&account).Error; err != nil {
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
	}

	accountInBytes, err := json.Marshal(account)
	if err != nil {
		http.Error(w, "Marshal failed: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := kafka.PushLedgerToQueue("initial_balance", accountInBytes); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(account); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
