package model

import (
	"time"
)

type PaymentSchedule struct {
	ID        int       `json:"id"         db:"id"`
	CreditID  int       `json:"credit_id"  db:"credit_id"`
	DueDate   time.Time `json:"due_date"   db:"due_date"`
	Amount    float64   `json:"amount"     db:"amount"`
	Paid      bool      `json:"paid"       db:"paid"`
	Penalty   float64   `json:"penalty"    db:"penalty"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
