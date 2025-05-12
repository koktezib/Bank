package service

import (
	"Bank/internal/model"
	"Bank/internal/repository"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrAccessDenied        = errors.New("access denied")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrUnsupportedCurrency = errors.New("unsupported currency; only RUB allowed")
)

type AccountService struct {
	db          *sql.DB
	userRepo    repository.UserRepository
	accountRepo repository.AccountRepository
	txRepo      repository.TransactionRepository
	mailSvc     MailService
}

func NewAccountService(
	db *sql.DB,
	ur repository.UserRepository,
	ar repository.AccountRepository,
	tr repository.TransactionRepository,
	mailSvc MailService,
) *AccountService {
	return &AccountService{
		db:          db,
		userRepo:    ur,
		accountRepo: ar,
		txRepo:      tr,
		mailSvc:     mailSvc,
	}
}

func (s *AccountService) CreateAccount(userID int, req *model.AccountCreate) (*model.Account, error) {
	if req.Currency != "RUB" {
		return nil, ErrUnsupportedCurrency
	}

	account := &model.Account{
		UserID:   userID,
		Balance:  0,
		Currency: req.Currency,
	}
	if err := s.accountRepo.Create(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) GetUserAccounts(userID int) ([]*model.Account, error) {
	return s.accountRepo.ListByUser(userID)
}

func (s *AccountService) Deposit(userID, accountID int, amount float64) (*model.Transaction, error) {
	acc, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, ErrAccessDenied
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	newBal := acc.Balance + amount
	if err := s.accountRepo.UpdateBalance(tx, accountID, newBal); err != nil {
		tx.Rollback()
		return nil, err
	}

	t := &model.Transaction{
		AccountID:   accountID,
		Amount:      amount,
		Type:        "deposit",
		Description: "Пополнение счёта",
	}
	if err := s.txRepo.CreateTx(tx, t); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(userID)
	if err == nil {
		subject := "Ваш счёт пополнен"
		body := fmt.Sprintf(
			"<h1>Пополнение счёта</h1><p>Сумма: <strong>%.2f RUB</strong></p>"+
				"<p>Новый баланс: <strong>%.2f RUB</strong></p>",
			amount, newBal,
		)
		_ = s.mailSvc.Send(user.Email, subject, body)
	}

	return t, nil
}

func (s *AccountService) Withdraw(userID, accountID int, amount float64) (*model.Transaction, error) {
	acc, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, ErrAccessDenied
	}
	if acc.Balance < amount {
		return nil, ErrInsufficientFunds
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	newBal := acc.Balance - amount
	if err = s.accountRepo.UpdateBalance(tx, accountID, newBal); err != nil {
		tx.Rollback()
		return nil, err
	}

	t := &model.Transaction{
		AccountID:   accountID,
		Amount:      amount,
		Type:        "withdraw",
		Description: "Снятие со счёта",
	}
	if err = s.txRepo.CreateTx(tx, t); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if user, e := s.userRepo.GetByID(userID); e == nil {
		subject := "Со счёта сняты средства"
		body := fmt.Sprintf(
			"<h1>Снятие со счёта</h1>"+
				"<p>Сумма: <strong>%.2f RUB</strong></p>"+
				"<p>Новый баланс: <strong>%.2f RUB</strong></p>",
			amount, newBal,
		)
		_ = s.mailSvc.Send(user.Email, subject, body)
	}

	return t, nil
}

func (s *AccountService) Transfer(userID, fromID, toID int, amount float64) (*model.Transaction, *model.Transaction, error) {
	fromAcc, err := s.accountRepo.GetByID(fromID)
	if err != nil {
		return nil, nil, err
	}
	if fromAcc.UserID != userID {
		return nil, nil, ErrAccessDenied
	}
	if fromAcc.Balance < amount {
		return nil, nil, ErrInsufficientFunds
	}

	toAcc, err := s.accountRepo.GetByID(toID)
	if err != nil {
		return nil, nil, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, nil, err
	}

	if err = s.accountRepo.UpdateBalance(tx, fromID, fromAcc.Balance-amount); err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	if err = s.accountRepo.UpdateBalance(tx, toID, toAcc.Balance+amount); err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tFrom := &model.Transaction{
		AccountID:   fromID,
		Amount:      amount,
		Type:        "transfer",
		Description: "to:" + strconv.Itoa(toID),
	}
	if err = s.txRepo.CreateTx(tx, tFrom); err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tTo := &model.Transaction{
		AccountID:   toID,
		Amount:      amount,
		Type:        "transfer",
		Description: "from:" + strconv.Itoa(fromID),
	}
	if err = s.txRepo.CreateTx(tx, tTo); err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	if user, e := s.userRepo.GetByID(userID); e == nil {
		subject := "Перевод отправлен"
		body := fmt.Sprintf(
			"<h1>Перевод</h1>"+
				"<p>Вы отправили <strong>%.2f RUB</strong> на счёт #%d</p>"+
				"<p>Ваш новый баланс: <strong>%.2f RUB</strong></p>",
			amount, toID, fromAcc.Balance-amount,
		)
		_ = s.mailSvc.Send(user.Email, subject, body)
	}
	if recipient, e := s.userRepo.GetByID(toAcc.UserID); e == nil {
		subject := "Вам поступил перевод"
		body := fmt.Sprintf(
			"<h1>Перевод</h1>"+
				"<p>На ваш счёт #%d поступило <strong>%.2f RUB</strong></p>"+
				"<p>Ваш новый баланс: <strong>%.2f RUB</strong></p>",
			toID, amount, toAcc.Balance+amount,
		)
		_ = s.mailSvc.Send(recipient.Email, subject, body)
	}

	return tFrom, tTo, nil
}
