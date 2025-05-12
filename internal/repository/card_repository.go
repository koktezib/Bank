package repository

import (
	"Bank/internal/model"
	"database/sql"
	"errors"
)

var ErrCardNotFound = errors.New("card not found")

type CardRepository interface {
	Create(c *model.Card) error
	ListByAccount(accountID int) ([]*model.Card, error)
	GetByID(id int) (*model.Card, error)
}

type cardRepo struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) CardRepository {
	return &cardRepo{db: db}
}

func (r *cardRepo) Create(c *model.Card) error {
	query := `
        INSERT INTO cards(account_id, number_encrypted, expiry_encrypted, cvv_hash, hmac)
        VALUES($1, $2, $3, $4, $5)
        RETURNING id, created_at
    `
	return r.db.QueryRow(query,
		c.AccountID, c.NumberEncrypted, c.ExpiryEncrypted, c.CVVHash, c.HMAC,
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *cardRepo) ListByAccount(accountID int) ([]*model.Card, error) {
	query := `
        SELECT id, account_id, number_encrypted, expiry_encrypted, cvv_hash, hmac, created_at
        FROM cards WHERE account_id = $1
    `
	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*model.Card
	for rows.Next() {
		c := &model.Card{}
		if err := rows.Scan(&c.ID, &c.AccountID, &c.NumberEncrypted, &c.ExpiryEncrypted, &c.CVVHash, &c.HMAC, &c.CreatedAt); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

func (r *cardRepo) GetByID(id int) (*model.Card, error) {
	c := &model.Card{}
	query := `
        SELECT id, account_id, number_encrypted, expiry_encrypted, cvv_hash, hmac, created_at
        FROM cards WHERE id = $1
    `
	err := r.db.QueryRow(query, id).
		Scan(&c.ID, &c.AccountID, &c.NumberEncrypted, &c.ExpiryEncrypted, &c.CVVHash, &c.HMAC, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrCardNotFound
	}
	return c, err
}
