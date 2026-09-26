package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Solomiam356/witness-backend/internal/middleware"
	"github.com/Solomiam356/witness-backend/internal/repository"
)

type ReflectionHandler struct {
	repo *repository.ReflectionRepository
}

func NewReflectionHandler(repo *repository.ReflectionRepository) *ReflectionHandler {
	return &ReflectionHandler{repo: repo}
}

func (h *ReflectionHandler) GetToday(w http.ResponseWriter, r *http.Request) {
	ref, text, err := h.repo.GetTodayVerse(r.Context())
	if err != nil {
		http.Error(w, "Не вдалося отримати вірш дня", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"verse_ref":  ref,
		"verse_text": text,
	})
}

func (h *ReflectionHandler) SaveMorning(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req struct {
		VerseRef  string `json:"verse_ref"`
		VerseText string `json:"verse_text"`
		Note      string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некоректний JSON", http.StatusBadRequest)
		return
	}

	if err := h.repo.SaveMorningNote(r.Context(), userID, req.VerseRef, req.VerseText, req.Note); err != nil {
		http.Error(w, "Не вдалося зберегти нотатку", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ранкову нотатку збережено"})
}

func (h *ReflectionHandler) SaveEvening(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req struct {
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некоректний JSON", http.StatusBadRequest)
		return
	}

	if err := h.repo.SaveEveningNote(r.Context(), userID, req.Note); err != nil {
		http.Error(w, "Не вдалося зберегти вечірню нотатку", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Вечірню нотатку збережено"})
}