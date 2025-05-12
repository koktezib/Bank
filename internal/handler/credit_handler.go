package handler

import (
	"Bank/internal/middleware"
	"Bank/internal/model"
	"Bank/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CreditHandler struct {
	creditSvc *service.CreditService
}

func NewCreditHandler(cs *service.CreditService) *CreditHandler {
	return &CreditHandler{creditSvc: cs}
}

func (h *CreditHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	var req model.CreditCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	credit, schedule, err := h.creditSvc.CreateCredit(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	// возвращаем сам кредит и сразу весь график
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"credit":   credit,
		"schedule": schedule,
	})
}

func (h *CreditHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))
	idStr := mux.Vars(r)["creditId"]
	creditID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid credit id", http.StatusBadRequest)
		return
	}

	sched, err := h.creditSvc.GetSchedule(userID, creditID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	json.NewEncoder(w).Encode(sched)
}
