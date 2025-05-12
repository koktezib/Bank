package handler

import (
	"Bank/internal/model"
	"Bank/internal/repository"
	"Bank/internal/service"
	"encoding/json"
	"errors"
	"net/http"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(a *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: a}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.authSvc.Register(&req)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			http.Error(w, "user already exists", http.StatusConflict)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.authSvc.Login(&req)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
