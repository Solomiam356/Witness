package handler

import (
	"strconv"
	"encoding/json"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/Solomiam356/witness-backend/internal/domain"
	"github.com/Solomiam356/witness-backend/internal/service"
	"github.com/Solomiam356/witness-backend/internal/middleware"
	"github.com/google/uuid"
)

type TestimonyHandler struct {
	srv *service.TestimonyService
}

func NewTestimonyHandler(srv *service.TestimonyService) *TestimonyHandler {
	return &TestimonyHandler{srv: srv}
}

// Create — створення нового свідчення
func (h *TestimonyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Користувач не авторизований", http.StatusUnauthorized)
		return
	}

	var t domain.Testimony
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Невалідний JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	t.UserID = userID

	if err := h.srv.CreateTestimony(r.Context(), &t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// GetAll — отримання всіх свідчень користувача
func (h *TestimonyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Користувач не авторизований", http.StatusUnauthorized)
		return
	}

	list, err := h.srv.GetTestimoniesByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Помилка отримання даних: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(list)
}

func (h *TestimonyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Користувач не авторизований", http.StatusUnauthorized)
		return
	}

	testimonyID := chi.URLParam(r, "id")
	if testimonyID == "" {
		http.Error(w, "ID свідчення не вказано", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(testimonyID); err != nil {
		w.Header().Set("Content-Type", "application/json")
	    w.WriteHeader(http.StatusBadRequest)
	    w.Write([]byte(`{"error": "Невалідний формат ID. Очікується UUID"}`))
		return
	}

	if err := h.srv.DeleteTestimony(r.Context(), testimonyID, userID); err != nil {
		http.Error(w, "Не вдалося видалити свідчення: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TestimonyHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")
	filterUserID := r.URL.Query().Get("user_id")

	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	result, err := h.srv.GetFeed(r.Context(), cursor, limit, search, filterUserID)
	if err != nil {
		http.Error(w, "Не вдалося завантажити стрічку: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

