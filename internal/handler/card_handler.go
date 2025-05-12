package handler

import (
	"Bank/internal/middleware"
	"Bank/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
)

type CardHandler struct {
	cardSvc *service.CardService
}

func NewCardHandler(c *service.CardService) *CardHandler {
	return &CardHandler{cardSvc: c}
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))
	accountID, err := strconv.Atoi(r.URL.Query().Get("account_id"))
	if err != nil {
		http.Error(w, "account_id required", http.StatusBadRequest)
		return
	}

	card, err := h.cardSvc.GenerateCard(userID, accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"card_id": card.ID,
	})
}

func (h *CardHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))
	cards, err := h.cardSvc.ListCards(userID)
	if err != nil {
		http.Error(w, "cannot list cards", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cards)
}
