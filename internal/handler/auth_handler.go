package handler

import (
	"encoding/json"
	"net/http"
	"github.com/Solomiam356/witness-backend/internal/service"
	"github.com/Solomiam356/witness-backend/internal/middleware"
)

type AuthHandler struct {
	srv *service.SessionService
}

func NewAuthHandler(srv *service.SessionService) *AuthHandler {
	return &AuthHandler{srv:srv}
}

func (h *AuthHandler) GetMySessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Користувач не авторизований", http.StatusUnauthorized)
		return
	}

	sessions, err := h.srv.GetUserSessions(r.Context(), userID)
	if err != nil {
		http.Error(w, "Помилка при отриманні сесій", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}