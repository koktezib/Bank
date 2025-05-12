package repository

import (
	"Bank/internal/model"
	"database/sql"
	"errors"
)

var ErrCreditNotFound = errors.New("credit not found")

type CreditRepository interface {
	Create(c *model.Credit) error
	GetByID(id int) (*model.Credit, error)
	ListByAccount(accountID int) ([]*model.Credit, error)
}

type creditRepo struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) CreditRepository {
	return &creditRepo{db: db}
}

func (r *creditRepo) Create(c *model.Credit) error {
	query := `
        INSERT INTO credits(account_id, principal, annual_rate, term_months)
        VALUES($1, $2, $3, $4)
        RETURNING id, created_at
    `
	return r.db.QueryRow(query, c.AccountID, c.Principal, c.AnnualRate, c.TermMonths).
		Scan(&c.ID, &c.CreatedAt)
}

func (r *creditRepo) GetByID(id int) (*model.Credit, error) {
	c := &model.Credit{}
	query := `
        SELECT id, account_id, principal, annual_rate, term_months, created_at
        FROM credits WHERE id = $1
    `
	err := r.db.QueryRow(query, id).
		Scan(&c.ID, &c.AccountID, &c.Principal, &c.AnnualRate, &c.TermMonths, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCreditNotFound
	}
	return c, err
}

func (r *creditRepo) ListByAccount(accountID int) ([]*model.Credit, error) {
	query := `
        SELECT id, account_id, principal, annual_rate, term_months, created_at
        FROM credits WHERE account_id = $1
    `
	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Credit
	for rows.Next() {
		c := &model.Credit{}
		if err := rows.Scan(&c.ID, &c.AccountID, &c.Principal, &c.AnnualRate, &c.TermMonths, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}
