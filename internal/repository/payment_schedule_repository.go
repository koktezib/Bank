package repository

import (
	"Bank/internal/model"
	"database/sql"
	"time"
)

type PaymentScheduleRepository interface {
	Create(ps *model.PaymentSchedule) error
	ListByCredit(creditID int) ([]*model.PaymentSchedule, error)
	MarkPaid(tx *sql.Tx, scheduleID int, penalty float64) error
	ListDue(date time.Time) ([]*model.PaymentSchedule, error)
	ListByAccountDueBetween(accountID int, from, to time.Time) ([]*model.PaymentSchedule, error)
}

type paymentScheduleRepo struct {
	db *sql.DB
}

func NewPaymentScheduleRepository(db *sql.DB) PaymentScheduleRepository {
	return &paymentScheduleRepo{db: db}
}

func (r *paymentScheduleRepo) Create(ps *model.PaymentSchedule) error {
	query := `
        INSERT INTO payment_schedules(credit_id, due_date, amount, paid, penalty)
        VALUES($1, $2, $3, $4, $5)
        RETURNING id, created_at
    `
	return r.db.QueryRow(query,
		ps.CreditID, ps.DueDate, ps.Amount, ps.Paid, ps.Penalty,
	).Scan(&ps.ID, &ps.CreatedAt)
}

func (r *paymentScheduleRepo) ListByCredit(creditID int) ([]*model.PaymentSchedule, error) {
	query := `
        SELECT id, credit_id, due_date, amount, paid, penalty, created_at
        FROM payment_schedules WHERE credit_id = $1 ORDER BY due_date
    `
	rows, err := r.db.Query(query, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.PaymentSchedule
	for rows.Next() {
		ps := &model.PaymentSchedule{}
		if err := rows.Scan(&ps.ID, &ps.CreditID, &ps.DueDate, &ps.Amount, &ps.Paid, &ps.Penalty, &ps.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, ps)
	}
	return list, rows.Err()
}

func (r *paymentScheduleRepo) MarkPaid(tx *sql.Tx, scheduleID int, penalty float64) error {
	query := `
        UPDATE payment_schedules
        SET paid = TRUE, penalty = $1
        WHERE id = $2
    `
	res, err := tx.Exec(query, penalty, scheduleID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *paymentScheduleRepo) ListDue(date time.Time) ([]*model.PaymentSchedule, error) {
	query := `
        SELECT id, credit_id, due_date, amount, paid, penalty, created_at
        FROM payment_schedules
        WHERE paid = FALSE AND due_date <= $1
    `
	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.PaymentSchedule
	for rows.Next() {
		ps := &model.PaymentSchedule{}
		if err := rows.Scan(&ps.ID, &ps.CreditID, &ps.DueDate, &ps.Amount, &ps.Paid, &ps.Penalty, &ps.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, ps)
	}
	return list, rows.Err()
}

func (r *paymentScheduleRepo) ListByAccountDueBetween(accountID int, from, to time.Time) ([]*model.PaymentSchedule, error) {
	query := `
        SELECT ps.id, ps.credit_id, ps.due_date, ps.amount, ps.paid, ps.penalty, ps.created_at
        FROM payment_schedules ps
        JOIN credits c ON ps.credit_id = c.id
        WHERE c.account_id = $1
          AND ps.due_date > $2 AND ps.due_date <= $3
    `
	rows, err := r.db.Query(query, accountID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*model.PaymentSchedule
	for rows.Next() {
		ps := &model.PaymentSchedule{}
		if err := rows.Scan(&ps.ID, &ps.CreditID, &ps.DueDate, &ps.Amount, &ps.Paid, &ps.Penalty, &ps.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, ps)
	}
	return out, rows.Err()
}
