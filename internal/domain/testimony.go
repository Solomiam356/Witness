package domain

import (
	"time"
	"github.com/lib/pq"
	)

type Testimony struct {
	ID string `json:"id"`
	UserID string `json:"user_id"`
	Title string `json:"title"`
	Content string `json:"content"`
	Summary string `json:"summary"`
	Tags pq.StringArray `gorm:"type:text[]" json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}