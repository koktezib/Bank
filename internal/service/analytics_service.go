// internal/service/analytics_service.go
package service

import (
	"Bank/internal/model"
	"Bank/internal/repository"
	"time"
)

type AnalyticsService struct {
	txRepo       repository.TransactionRepository
	acctRepo     repository.AccountRepository
	scheduleRepo repository.PaymentScheduleRepository
}

func NewAnalyticsService(
	tr repository.TransactionRepository,
	ar repository.AccountRepository,
	sr repository.PaymentScheduleRepository,
) *AnalyticsService {
	return &AnalyticsService{txRepo: tr, acctRepo: ar, scheduleRepo: sr}
}

func (s *AnalyticsService) GetMonthlyStats(userID int) (*model.MonthlyStats, error) {
	accounts, err := s.acctRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()

	var totalIncome, totalExpense, totalCreditLoad float64

	for _, acc := range accounts {
		txs, err := s.txRepo.ListByAccountBetween(acc.ID, from, to)
		if err != nil {
			return nil, err
		}
		for _, t := range txs {
			if t.Type == "deposit" || t.Type == "transfer_in" {
				totalIncome += t.Amount
			} else {
				totalExpense += t.Amount
			}
		}
		scheds, err := s.scheduleRepo.ListByAccountDueBetween(acc.ID, time.Time{}, time.Now().AddDate(10, 0, 0))
		if err != nil {
			return nil, err
		}
		for _, ps := range scheds {
			if !ps.Paid {
				totalCreditLoad += ps.Amount
			}
		}
	}

	return &model.MonthlyStats{
		Income:     totalIncome,
		Expense:    totalExpense,
		CreditLoad: totalCreditLoad,
	}, nil
}

func (s *AnalyticsService) GetBalanceForecast(userID, accountID, days int) ([]*model.BalanceForecast, error) {
	acc, err := s.acctRepo.GetByID(accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, ErrAccessDenied
	}

	now := time.Now()
	future := now.AddDate(0, 0, days)
	scheds, err := s.scheduleRepo.ListByAccountDueBetween(accountID, now, future)
	if err != nil {
		return nil, err
	}

	payMap := make(map[string]float64)
	for _, ps := range scheds {
		key := ps.DueDate.Format("2006-01-02")
		payMap[key] += ps.Amount + ps.Penalty
	}

	var result []*model.BalanceForecast
	bal := acc.Balance
	for i := 1; i <= days; i++ {
		date := now.AddDate(0, 0, i)
		key := date.Format("2006-01-02")
		if amt, ok := payMap[key]; ok {
			bal -= amt
		}
		result = append(result, &model.BalanceForecast{
			Date:    date,
			Balance: bal,
		})
	}
	return result, nil
}
