package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Solomiam356/witness-backend/internal/middleware"
	"github.com/Solomiam356/witness-backend/internal/repository"
	"github.com/Solomiam356/witness-backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

type ReportHandler struct {
	repo *repository.ReportRepository
}

func NewReportHandler(repo *repository.ReportRepository) *ReportHandler {
	return &ReportHandler{repo: repo}
}

type CreateReportRequest struct {
	Reason string `json:"reason"`
}

func (h *ReportHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Користувач не авторизований", http.StatusUnauthorized)
		return
	}

	testimonyID := chi.URLParam(r, "id")
	if testimonyID == "" {
		http.Error(w, "ID свідоцтва не вказано", http.StatusBadRequest)
		return
	}

	var req CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Невірний формат даних", http.StatusBadRequest)
		return
	}

	report := &domain.Report{
		ReporterID:      userID,
		TestimonyID: testimonyID,
		Reason:      req.Reason,
	}

	if err := h.repo.Create(r.Context(), report); err != nil {
		http.Error(w, "Помилка при створенні скарги: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Скарга успішно надіслана на розгляд"}`))
}