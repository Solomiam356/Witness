package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/Solomiam356/witness-backend/internal/domain"
	"github.com/Solomiam356/witness-backend/internal/repository"
)

type TestimonyService struct {
	repo  *repository.TestimonyRepository
	aiSvc *AIService
}

func NewTestimonyService(repo *repository.TestimonyRepository, aiSvc *AIService) *TestimonyService {
	return &TestimonyService{repo: repo, aiSvc: aiSvc}
}

func (s *TestimonyService) CreateTestimony(ctx context.Context, t *domain.Testimony) error {
	if strings.TrimSpace(t.Title) == "" {
		return errors.New("заголовок свідчення не може бути порожнім")
	}
	if strings.TrimSpace(t.Content) == "" {
		return errors.New("текст свідчення не може бути порожнім")
	}

	// 1. Зберігаємо свідчення в БД з is_published = false
	err := s.repo.Create(ctx, t)
	if err != nil {
		return fmt.Errorf("помилка збереження свідчення: %w", err)
	}

	// 2. Асинхронна модерація та генерація тегів у фоновій горутині
	go func(id string, content string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		analysis, err := s.aiSvc.AnalyzeAndSummarize(bgCtx, content)
		if err != nil {
			log.Printf("[AI MODERATION ERROR] Не вдалося проаналізувати свідчення %s: %v", id, err)
			return
		}

		if !analysis.IsSafe {
			log.Printf("[AI MODERATION REJECTED] Свідчення %s відхилено (небезпечний контент)", id)
			_ = s.repo.HardDelete(bgCtx, id)
			return
		}

		// Публікуємо свідчення (is_published = true) та оновлюємо summary і теги
		tags := pq.StringArray(analysis.Tags)
		err = s.repo.UpdateModerationStatus(bgCtx, id, analysis.Summary, tags, true)
		if err != nil {
			log.Printf("[AI MODERATION ERROR] Помилка публікації свідчення %s: %v", id, err)
		} else {
			log.Printf("[AI MODERATION SUCCESS] Свідчення %s успішно перевірено та опубліковано", id)
		}
	}(t.ID, t.Content)

	return nil
}

func (s *TestimonyService) GetTestimoniesByUserID(ctx context.Context, userID string) ([]domain.Testimony, error) {
	return s.repo.GetAllByUserID(ctx, userID)
}

func (s *TestimonyService) DeleteTestimony(ctx context.Context, id string, userID string) error {
	return s.repo.DeleteByID(ctx, id, userID)
}

type PaginatedTestimonies struct {
	Data       []domain.Testimony `json:"data"`
	NextCursor string             `json:"next_cursor"`
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
		Data:       list,
		NextCursor: nextCursor,
	}, nil
}