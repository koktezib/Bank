package model

import (
	"time"
)

type Account struct {
	ID        int       `json:"id"       db:"id"`
	UserID    int       `json:"user_id"  db:"user_id"`
	Balance   float64   `json:"balance"  db:"balance"`
	Currency  string    `json:"currency" db:"currency"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AccountCreate struct {
	Currency string `json:"currency" validate:"required,eq=RUB"`
}

func (a *AccountCreate) Validate() error {
	return validate.Struct(a)
}
