package models

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"math"
)

type Account struct {
	ID               uint          `gorm:"primaryKey;autoIncrement;not null" json:"id"`
	Name             string        `gorm:"not null" validate:"required" json:"name"`
	InitialBalance   int64         `gorm:"not null" validate:"required" json:"initial_balance"`
	TransactionsFrom []Transaction `gorm:"foreignKey:FromID" json:"transactions_from,omitempty"`
	TransactionsTo   []Transaction `gorm:"foreignKey:ToID" json:"transactions_to,omitempty"`
}

func ValidateAccount(account *Account) error {
	validate := validator.New()
	return validate.Struct(account)
}

func (a *Account) SetInitialBalanceFromDecimal(decimalBalance float64) error {
	if decimalBalance < 0 {
		return errors.New("balance cannot be negative")
	}
	a.InitialBalance = int64(math.Round(decimalBalance * 1_000_000)) // Convert to micro-units
	return nil
}
