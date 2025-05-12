package model

import (
	"time"
)

type Credit struct {
	ID         int       `json:"id"           db:"id"`
	AccountID  int       `json:"account_id"   db:"account_id"`
	Principal  float64   `json:"principal"    db:"principal"`
	AnnualRate float64   `json:"annual_rate"  db:"annual_rate"`
	TermMonths int       `json:"term_months"  db:"term_months"`
	CreatedAt  time.Time `json:"created_at"   db:"created_at"`
}

type CreditCreate struct {
	AccountID  int     `json:"account_id"  validate:"required"`
	Principal  float64 `json:"principal"   validate:"required,gt=0"`
	TermMonths int     `json:"term_months" validate:"required,gt=0"`
}

func (c *CreditCreate) Validate() error {
	return validate.Struct(c)
}
