package database

import (
	"context"
	"log"
)

func InitTables() error {
	// 1. Зносимо стару версію таблиці, яка не мала поля user_id
	_, err := DB.Exec(context.Background(), "DROP TABLE IF EXISTS tasks;")
	if err != nil {
		return err
	}

	// 2. Створюємо правильну таблицю під Witness V2.1
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL,
		title TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = DB.Exec(context.Background(), query)
	if err != nil {
		return err
	}

	log.Println("Таблиця 'tasks' успішно перевизначена та синхронізована з архітектурою V2.1")
	return nil
}