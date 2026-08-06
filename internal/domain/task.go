package domain

import "time"

type Task struct {
	ID string `json:"id"`
	UserID string `json:"user_id"`
	Title string `json:"title"`
	Description string `json:"description"`
	Status string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}