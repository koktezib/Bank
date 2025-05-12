// internal/handler/analytics_handler.go
package handler

import (
	"Bank/internal/middleware"
	"Bank/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type AnalyticsHandler struct {
	analyticsSvc *service.AnalyticsService
}

func NewAnalyticsHandler(s *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsSvc: s}
}

func (h *AnalyticsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))
	stats, err := h.analyticsSvc.GetMonthlyStats(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(stats)
}

func (h *AnalyticsHandler) Predict(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))
	accountID, _ := strconv.Atoi(mux.Vars(r)["accountId"])
	days, err := strconv.Atoi(r.URL.Query().Get("days"))
	if err != nil || days < 1 {
		days = 30 // default
	}

	forecast, err := h.analyticsSvc.GetBalanceForecast(userID, accountID, days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(forecast)
}
