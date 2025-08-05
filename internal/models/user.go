package models

type User struct {
	ID        string `db:"id"`
	TelegramID string `db:"telegram_id"`
	Password  string `db:"password"`
}
