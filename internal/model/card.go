package model

import (
	"time"
)

type Card struct {
	ID              int       `json:"id"                 db:"id"`
	AccountID       int       `json:"account_id"         db:"account_id"`
	NumberEncrypted []byte    `json:"-"                  db:"number_encrypted"`
	ExpiryEncrypted []byte    `json:"-"                  db:"expiry_encrypted"`
	CVVHash         string    `json:"-"                  db:"cvv_hash"`
	HMAC            string    `json:"-"                  db:"hmac"`
	CreatedAt       time.Time `json:"created_at"         db:"created_at"`
}

type CardResponse struct {
	ID        int       `json:"id"`
	Number    string    `json:"number"`
	Expiry    string    `json:"expiry"`
	CreatedAt time.Time `json:"created_at"`
}
type CardCreate struct {
	AccountID int `json:"account_id" validate:"required"`
}

func (c *CardCreate) Validate() error {
	return validate.Struct(c)
}
