package repository

import (
	"fmt"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Solomiam356/witness-backend/internal/domain"
)

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	if task.Status == "" {
		task.Status = "pending"
	}
	
	query := `
		INSERT INTO tasks (id, user_id, title, description, status, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`

	return r.db.QueryRow(ctx, query, task.UserID, task.Title, task.Description, task.Status).Scan(&task.ID, &task.CreatedAt)
}

func (r *TaskRepository) GetAllByUserID(ctx context.Context, userID string) ([]domain.Task, error) {
	query := `SELECT id, user_id, title, description, status, created_at
			  FROM tasks
	          WHERE user_id = $1
			  ORDER BY created_at DESC`

 	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task

	for rows.Next() {
		var t domain.Task
		err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
	
		tasks = append(tasks, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id string, userID string, status  string) error {
query := `UPDATE tasks SET status = $1 WHERE id = $2 AND user_id = $3`

result, err := r.db.Exec(ctx, query, status, id, userID)
if err != nil {
	return err
}

if result.RowsAffected() == 0 {
	return fmt.Errorf("завдання не знайдено або у вас не має доступу")
}

return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string, userID string) error {
	query := `DELETE  FROM tasks WHERE id = $1 AND user_id = $2`

	res, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("завдання не знайдено або у вас немає прав на його видалення")
	}

	return nil
}