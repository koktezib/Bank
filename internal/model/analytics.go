package model

import "time"

type MonthlyStats struct {
	Income     float64 `json:"income"`
	Expense    float64 `json:"expense"`
	CreditLoad float64 `json:"credit_load"`
}

type BalanceForecast struct {
	Date    time.Time `json:"date"`
	Balance float64   `json:"balance"`
}
