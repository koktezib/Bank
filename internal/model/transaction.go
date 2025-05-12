package model

import (
	"time"
)

type Transaction struct {
	ID          int       `json:"id"          db:"id"`
	AccountID   int       `json:"account_id"  db:"account_id"`
	Amount      float64   `json:"amount"      db:"amount"`
	Type        string    `json:"type"        db:"type"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
}

type TransactionCreate struct {
	AccountID   int     `json:"account_id" validate:"required"`
	Amount      float64 `json:"amount"     validate:"required,gt=0"`
	Type        string  `json:"type"       validate:"required,oneof=deposit withdraw transfer"`
	Description string  `json:"description"`
}

func (t *TransactionCreate) Validate() error {
	return validate.Struct(t)
}
