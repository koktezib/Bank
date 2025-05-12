package handler

import (
	"Bank/internal/middleware"
	"Bank/internal/model"
	"Bank/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	_ "github.com/gorilla/mux"
)

type AccountHandler struct {
	accSvc *service.AccountService
}

func NewAccountHandler(s *service.AccountService) *AccountHandler {
	return &AccountHandler{accSvc: s}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	var req model.AccountCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	acc, err := h.accSvc.CreateAccount(userID, &req)
	if err != nil {
		http.Error(w, "cannot create account", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(acc)
}

func (h *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))
	list, err := h.accSvc.GetUserAccounts(userID)
	if err != nil {
		http.Error(w, "cannot fetch accounts", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

type AmountRequest struct {
	AccountID int     `json:"account_id" validate:"required"`
	Amount    float64 `json:"amount"     validate:"required,gt=0"`
}

func (ar *AmountRequest) Validate() error { return model.ValidateStruct(ar) }

func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	var req AmountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.accSvc.Deposit(userID, req.AccountID, req.Amount)
	if err != nil {
		code := http.StatusInternalServerError
		if err == service.ErrAccessDenied {
			code = http.StatusForbidden
		}
		http.Error(w, err.Error(), code)
		return
	}
	json.NewEncoder(w).Encode(tx)
}

func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	var req AmountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.accSvc.Withdraw(userID, req.AccountID, req.Amount)
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case service.ErrAccessDenied:
			code = http.StatusForbidden
		case service.ErrInsufficientFunds:
			code = http.StatusConflict
		}
		http.Error(w, err.Error(), code)
		return
	}
	json.NewEncoder(w).Encode(tx)
}

type TransferRequest struct {
	FromAccountID int     `json:"from_account_id" validate:"required"`
	ToAccountID   int     `json:"to_account_id"   validate:"required"`
	Amount        float64 `json:"amount"          validate:"required,gt=0"`
}

func (tr *TransferRequest) Validate() error { return model.ValidateStruct(tr) }

func (h *AccountHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	txFrom, txTo, err := h.accSvc.Transfer(userID, req.FromAccountID, req.ToAccountID, req.Amount)
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case service.ErrAccessDenied:
			code = http.StatusForbidden
		case service.ErrInsufficientFunds:
			code = http.StatusConflict
		}
		http.Error(w, err.Error(), code)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"debit":  txFrom,
		"credit": txTo,
	})
}
