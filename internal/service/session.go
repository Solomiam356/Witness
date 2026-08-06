package service

import (
	"context"
	"github.com/Solomiam356/witness-backend/internal/domain"
	"github.com/Solomiam356/witness-backend/internal/repository"
)

type SessionService struct {
	repo *repository.SessionRepository
}

func NewSessionService(repo *repository.SessionRepository) *SessionService {
	return &SessionService{repo:repo}
}

func (s *SessionService) GetUserSessions(ctx context.Context, userID string) ([]domain.Session, error) {
	return s.repo.GetSessionsByUserID(ctx, userID)
}