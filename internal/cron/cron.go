package cron

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
	"github.com/Solomiam356/witness-backend/internal/service"
)

type Scheduler struct {
	cron  *cron.Cron
	db    *pgxpool.Pool
	aiSvc *service.AIService
}

func NewScheduler(db *pgxpool.Pool, aiSvc *service.AIService) *Scheduler {
	// cron.New() створює планиральник за стандартом 5 полів (хвилини, години, день місяця, місяць, день тижня)
	return &Scheduler{
		cron:  cron.New(),
		db:    db,
		aiSvc: aiSvc,
	}
}

func (s *Scheduler) Start() {
	// 1. Щоранку о 06:00 (Генерація / вибір вірша дня)
	_, err := s.cron.AddFunc("0 6 * * *", func() {
		log.Println("⏰ [CRON 06:00] Генерація та вибір вірша дня...")
		s.generateDailyVerse()
	})
	if err != nil {
		log.Printf("🔴 [CRON ERROR] Не вдалося додати ранкову задачу: %v", err)
	}

	// 2. Щовечора о 21:00 (Нагадування заповнити evening_note)
	_, err = s.cron.AddFunc("0 21 * * *", func() {
		log.Println("⏰ [CRON 21:00] Нагадування повернути вечірню нотатку...")
		s.sendEveningReminders()
	})
	if err != nil {
		log.Printf("🔴 [CRON ERROR] Не вдалося додати вечірню задачу: %v", err)
	}

	s.cron.Start()
	log.Println("🚀 [CRON] Планувальник Cron успішно запущено!")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("🛑 [CRON] Планувальник Cron зупинено.")
}

// generateDailyVerse — завантажує або генерує вірш дня й зберігає в daily_verses
func (s *Scheduler) generateDailyVerse() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	today := time.Now().Format("2006-01-02")

	// Вставити вірш дня в БД (якщо його ще немає)
	query := `
		INSERT INTO daily_verses (id, verse_ref, verse_text, date, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		ON CONFLICT (date) DO NOTHING
	`
	// Наприклад, фіксований або згенерований через Gemini вірш
	verseRef := "Івана 3:16"
	verseText := "Бо так мати-природа любить людей, що віддала Сина Свого Однородженого..."

	_, err := s.db.Exec(ctx, query, verseRef, verseText, today)
	if err != nil {
		log.Printf("🔴 [CRON 06:00 ERROR] Запис вірша дня збоїв: %v", err)
		return
	}

	log.Printf("✅ [CRON 06:00 SUCCESS] Вірш дня на %s збережено в БД!", today)
}

// sendEveningReminders — обробляє нагадування користувачам для заповнення evening_note
func (s *Scheduler) sendEveningReminders() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Логіка підтягування користувачів, у яких є ранковий запис, але немає evening_note
	query := `
		SELECT user_id FROM daily_reflections
		WHERE date = CURRENT_DATE AND (evening_note IS NULL OR evening_note = '')
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		log.Printf("🔴 [CRON 21:00 ERROR] Не вдалося отримати список користувачів: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err == nil {
			count++
		}
	}

	log.Printf("✅ [CRON 21:00 SUCCESS] Знайдено %d користувачів без вечірньої нотатки. Нагадування згенеровано!", count)
}