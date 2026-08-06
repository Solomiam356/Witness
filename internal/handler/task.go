package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
	"github.com/Solomiam356/witness-backend/internal/domain"
	"github.com/Solomiam356/witness-backend/internal/service"
	"github.com/Solomiam356/witness-backend/internal/middleware"
)

type TaskHandler struct {
	srv *service.TaskService
}

func NewTaskHandler(srv *service.TaskService) *TaskHandler {
	return &TaskHandler{srv: srv}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Користувач не авторизований", http.StatusUnauthorized)
		return
	}

	var task domain.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Невалідний JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	task.UserID = userID

	 err := h.srv.CreateTask(r.Context(), &task)
	if err != nil {
		if strings.Contains(err.Error(), "порожнім") {
		http.Error(w, err.Error(), http.StatusBadRequest) 
		return
	}
	http.Error(w, "Внутрішня помилка: "+err.Error(), http.StatusInternalServerError)
	return
}
		
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Користувач не авторизований", http.StatusUnauthorized)
		return
	}

	list, err := h.srv.GetTaskByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Помилка завантаження тасок: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(list)
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Користувач не авторизований", http.StatusUnauthorized)
		return
	}

	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "ID завдання не вказано", http.StatusBadRequest)
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некоректний формат JSON", http.StatusBadRequest)
		return
	}

	if err := h.srv.UpdateTaskStatus(r.Context(), taskID, userID, req.Status); err != nil {
		http.Error(w, "Не вдалося оновити статус: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Статус завдання успішно оновлено!"}`))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Користувач не авторизований", http.StatusUnauthorized)
		return
	}

	err := h.srv.DeleteTask(r.Context(), id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Завдання успішно видалено!"}`))
}