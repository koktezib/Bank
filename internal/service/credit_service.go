// internal/service/credit_service.go
package service

import (
	"Bank/internal/model"
	"Bank/internal/repository"
	"database/sql"
	"errors"
	"math"
	"time"
)

var (
	ErrCreditNotYours = errors.New("credit does not belong to user")
)

type CreditService struct {
	db           *sql.DB
	creditRepo   repository.CreditRepository
	scheduleRepo repository.PaymentScheduleRepository
	accountRepo  repository.AccountRepository
	cbr          *CBRService
}

func NewCreditService(
	db *sql.DB,
	cr repository.CreditRepository,
	sr repository.PaymentScheduleRepository,
	ar repository.AccountRepository,
	cbrSvc *CBRService,
) *CreditService {
	return &CreditService{
		db:           db,
		creditRepo:   cr,
		scheduleRepo: sr,
		accountRepo:  ar,
		cbr:          cbrSvc,
	}
}

func (s *CreditService) CreateCredit(userID int, req *model.CreditCreate) (*model.Credit, []*model.PaymentSchedule, error) {
	acc, err := s.accountRepo.GetByID(req.AccountID)
	if err != nil {
		return nil, nil, err
	}
	if acc.UserID != userID {
		return nil, nil, ErrCreditNotYours
	}

	rate, err := s.cbr.GetRate()
	if err != nil {
		return nil, nil, err
	}

	monthlyRate := rate / 100.0 / 12.0

	n := float64(req.TermMonths)
	factor := (monthlyRate * math.Pow(1+monthlyRate, n)) /
		(math.Pow(1+monthlyRate, n) - 1)
	annuity := math.Round(req.Principal*factor*100) / 100

	credit := &model.Credit{
		AccountID:  req.AccountID,
		Principal:  req.Principal,
		AnnualRate: rate,
		TermMonths: req.TermMonths,
	}
	if err := s.creditRepo.Create(credit); err != nil {
		return nil, nil, err
	}

	schedules := make([]*model.PaymentSchedule, 0, req.TermMonths)
	outstanding := req.Principal

	for i := 1; i <= req.TermMonths; i++ {
		interest := math.Round(outstanding*monthlyRate*100) / 100
		principalPortion := math.Round((annuity-interest)*100) / 100

		payment := annuity
		if i == req.TermMonths {
			principalPortion = outstanding
			payment = math.Round((interest+principalPortion)*100) / 100
		}

		outstanding = math.Round((outstanding-principalPortion)*100) / 100

		due := time.Now().AddDate(0, i, 0)
		ps := &model.PaymentSchedule{
			CreditID: credit.ID,
			DueDate:  time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, due.Location()),
			Amount:   payment,
			Paid:     false,
			Penalty:  0,
		}
		if err := s.scheduleRepo.Create(ps); err != nil {
			return nil, nil, err
		}
		schedules = append(schedules, ps)
	}

	return credit, schedules, nil
}

func (s *CreditService) GetSchedule(userID, creditID int) ([]*model.PaymentSchedule, error) {
	cr, err := s.creditRepo.GetByID(creditID)
	if err != nil {
		return nil, err
	}
	acc, err := s.accountRepo.GetByID(cr.AccountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, ErrCreditNotYours
	}
	return s.scheduleRepo.ListByCredit(creditID)
}

func (s *CreditService) ProcessDuePayments() error {
	schedules, err := s.scheduleRepo.ListDue(time.Now())
	if err != nil {
		return err
	}

	for _, ps := range schedules {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}

		acc, err := s.accountRepo.GetByID(ps.CreditID) // получить связанный счёт
		if err != nil {
			tx.Rollback()
			continue
		}
		if acc.Balance >= ps.Amount {
			newBal := acc.Balance - ps.Amount
			if err = s.accountRepo.UpdateBalance(tx, acc.ID, newBal); err != nil {
				tx.Rollback()
				continue
			}
			if err = s.scheduleRepo.MarkPaid(tx, ps.ID, 0); err != nil {
				tx.Rollback()
				continue
			}
		} else {
			penalty := math.Round(ps.Amount*0.1*100) / 100
			if err = s.scheduleRepo.MarkPaid(tx, ps.ID, penalty); err != nil {
				tx.Rollback()
				continue
			}
		}

		tx.Commit()
	}
	return nil
}
