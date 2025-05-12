package repository

import (
	"Bank/internal/model"
	"database/sql"
	"errors"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountRepository interface {
	Create(a *model.Account) error
	GetByID(id int) (*model.Account, error)
	ListByUser(userID int) ([]*model.Account, error)
	UpdateBalance(tx *sql.Tx, accountID int, newBalance float64) error
}

type accountRepo struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) Create(a *model.Account) error {
	query := `
        INSERT INTO accounts(user_id, balance, currency)
        VALUES($1, $2, $3)
        RETURNING id, created_at
    `
	return r.db.QueryRow(query, a.UserID, a.Balance, a.Currency).
		Scan(&a.ID, &a.CreatedAt)
}

func (r *accountRepo) GetByID(id int) (*model.Account, error) {
	a := &model.Account{}
	query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE id = $1`
	err := r.db.QueryRow(query, id).
		Scan(&a.ID, &a.UserID, &a.Balance, &a.Currency, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrAccountNotFound
	}
	return a, err
}

func (r *accountRepo) ListByUser(userID int) ([]*model.Account, error) {
	query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Account
	for rows.Next() {
		a := &model.Account{}
		if err := rows.Scan(&a.ID, &a.UserID, &a.Balance, &a.Currency, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *accountRepo) UpdateBalance(tx *sql.Tx, accountID int, newBalance float64) error {
	query := `UPDATE accounts SET balance = $1 WHERE id = $2`
	res, err := tx.Exec(query, newBalance, accountID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrAccountNotFound
	}
	return nil
}
