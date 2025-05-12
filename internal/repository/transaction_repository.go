package repository

import (
	"Bank/internal/model"
	"database/sql"
	"time"
)

type TransactionRepository interface {
	CreateTx(tx *sql.Tx, t *model.Transaction) error
	ListByAccount(accountID int) ([]*model.Transaction, error)
	ListByAccountBetween(accountID int, from, to time.Time) ([]*model.Transaction, error)
}

type transactionRepo struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) CreateTx(tx *sql.Tx, t *model.Transaction) error {
	query := `
        INSERT INTO transactions(account_id, amount, type, description)
        VALUES($1, $2, $3, $4)
        RETURNING id, created_at
    `
	return tx.QueryRow(query, t.AccountID, t.Amount, t.Type, t.Description).
		Scan(&t.ID, &t.CreatedAt)
}

func (r *transactionRepo) ListByAccount(accountID int) ([]*model.Transaction, error) {
	query := `
        SELECT id, account_id, amount, type, description, created_at
        FROM transactions WHERE account_id = $1 ORDER BY created_at DESC
    `
	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Transaction
	for rows.Next() {
		t := &model.Transaction{}
		if err := rows.Scan(&t.ID, &t.AccountID, &t.Amount, &t.Type, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *transactionRepo) ListByAccountBetween(accountID int, from, to time.Time) ([]*model.Transaction, error) {
	query := `
        SELECT id, account_id, amount, type, description, created_at
        FROM transactions
        WHERE account_id = $1 AND created_at >= $2 AND created_at <= $3
    `
	rows, err := r.db.Query(query, accountID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*model.Transaction
	for rows.Next() {
		t := &model.Transaction{}
		if err := rows.Scan(&t.ID, &t.AccountID, &t.Amount, &t.Type, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
