package models

type Balance struct {
	IdAccount string `gorm:"primaryKey;not null" validate:"required" json:"id"`
	Balance   int64  `gorm:"not null" validate:"required" json:"initial_balance"`
}
