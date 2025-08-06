package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email      string    `json:"email" gorm:"column:email;uniqueIndex"`
	TelegramID *int64    `json:"telegram_id,omitempty" gorm:"column:telegram_id;uniqueIndex"`
	FirstName  string    `json:"first_name" gorm:"column:first_name"`
	LastName   string    `json:"last_name" gorm:"column:last_name"`
	Username   string    `json:"username" gorm:"column:username"`
	IsActive   bool      `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}