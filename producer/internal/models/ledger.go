package models

import (
	"github.com/go-playground/validator/v10"
	"time"
)

type Transaction struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;not null" json:"id"`
	FromID    string    `gorm:"not null" validate:"required" json:"from_id"`
	From      *Account  `gorm:"foreignKey:FromID;constraint:OnDelete:CASCADE" json:"-" validate:"-"`
	ToID      string    `gorm:"not null" validate:"required" json:"to_id"`
	To        *Account  `gorm:"foreignKey:ToID;constraint:OnDelete:CASCADE" json:"-" validate:"-"`
	Amount    int64     `gorm:"not null" validate:"required,gt=0" json:"amount"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
}

func ValidateLedger(ledger *Transaction) error {
	validate := validator.New()
	return validate.Struct(ledger)
}
