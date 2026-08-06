package service

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"
	"fmt"

    "github.com/lib/pq" 
	"github.com/Solomiam356/witness-backend/internal/domain"
	"github.com/Solomiam356/witness-backend/internal/repository"
)


type TestimonyService struct { 
	repo *repository.TestimonyRepository
	aiSvc *AIService
}

func NewTestimonyService(repo *repository.TestimonyRepository, aiSvc *AIService) *TestimonyService {
	return &TestimonyService{repo: repo, aiSvc: aiSvc}
}

func (s *TestimonyService) CreateTestimony(ctx context.Context, t *domain.Testimony)  error {
	if strings.TrimSpace(t.Title) == "" {
		return errors.New("заголовок свічення не може бути порожнім")
	}
	if strings.TrimSpace(t.Content) == "" {
		return errors.New("текст свідчення не може бути порожнім")
	}

	analysis, err := s.aiSvc.AnalyzeAndSummarize(ctx, t.Content)
	if err != nil {
		return fmt.Errorf("помилка автоматичної модерації: %w ", err)
	}

	if !analysis.IsSafe {
		return errors.New("свідчення не пройшло автоматичну модерацію (виявлено спам, нецензурну лексику або агресію)")
	}

	t.Summary = analysis.Summary
	t.Tags = pq.StringArray(analysis.Tags)

	return s.repo.Create(ctx, t)
}

func (s *TestimonyService) GetTestimoniesByUserID(ctx context.Context, userID string) ([]domain.Testimony, error) {
	return s.repo.GetAllByUserID(ctx, userID)
}

func (s *TestimonyService) DeleteTestimony(ctx context.Context, id string, userID string) error {
	return s.repo.DeleteByID(ctx, id, userID)
}

type PaginatedTestimonies struct {
	Data []domain.Testimony `json:"data"`
	NextCursor string `json:"next_cursor"`
}

func (s *TestimonyService) GetFeed(ctx context.Context, base64Cursor string, limit int, search string, filterUserID string) (*PaginatedTestimonies, error) {
	realCursor := ""
	if base64Cursor != "" {
		decodedBytes, err := base64.StdEncoding.DecodeString(base64Cursor)
		if err == nil {
			realCursor = string(decodedBytes)
		}
	}
	list, err := s.repo.GetFeed(ctx, realCursor, limit, search, filterUserID)
	if err != nil {
		return nil, err
	}

	nextCursor := ""
	 
	if len(list) > 0 {
		lastItem := list[len(list)-1]
		timeStr := lastItem.CreatedAt.Format(time.RFC3339Nano)
		nextCursor = base64.StdEncoding.EncodeToString([]byte(timeStr))
	}

	return &PaginatedTestimonies{
		Data: list,
		NextCursor: nextCursor,
	}, nil
}