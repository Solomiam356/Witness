package service

import (
	"fmt"
	"context"
	"strings"
	"github.com/Solomiam356/witness-backend/internal/domain"
	"github.com/Solomiam356/witness-backend/internal/repository"
)


type TaskService struct { 
	repo *repository.TaskRepository
}

func NewTaskService(repo *repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(ctx context.Context, task *domain.Task)   error {
	if strings.TrimSpace(task.Title) == "" {
	return fmt.Errorf("заголовок завдання не може бути порожнім")
	}
	return s.repo.Create(ctx, task)
	}


func (s *TaskService) GetTaskByUserID(ctx context.Context, userID string) ([]domain.Task, error) {
	return s.repo.GetAllByUserID(ctx, userID)
}

func (s *TaskService) UpdateTaskStatus(ctx context.Context, id string, userID string, status string)  error{
	if status != "pending" && status != "completed" {
		return fmt.Errorf("недопустимий статус: дозволено тільки 'pending' або 'completed'")
	}

	return s.repo.UpdateStatus(ctx, id, userID, status)
}

func (s *TaskService) DeleteTask(ctx context.Context, id string, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}
