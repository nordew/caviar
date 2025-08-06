package storage

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"caviar/internal/models"
	"caviar/pkg/apperror"
)

type userStorage struct {
	db *gorm.DB
}

func NewUserStorage(db *gorm.DB) *userStorage {
	return &userStorage{
		db: db.Session(&gorm.Session{
			PrepareStmt: true,
		}),
	}
}

func (s *userStorage) GetAllTelegramIDs(ctx context.Context) ([]string, error) {
	var telegramIDs []string
	
	err := s.db.WithContext(ctx).
		Model(&models.User{}).
		Pluck("telegram_id", &telegramIDs).
		Error
	
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to fetch telegram IDs")
	}
	
	return telegramIDs, nil
}

func (s *userStorage) GetTelegramIDByUserID(ctx context.Context, userID string) (string, error) {
	var telegramID string
	
	err := s.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Pluck("telegram_id", &telegramID).
		Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", apperror.New(
			apperror.CodeNotFound,
			fmt.Sprintf("user %s not found", userID),
		)
	}
	
	if err != nil {
		return "", apperror.Wrap(err, apperror.CodeInternal, "failed to fetch telegram ID")
	}
	
	if telegramID == "" {
		return "", apperror.New(
			apperror.CodeNotFound,
			fmt.Sprintf("telegram ID not found for user %s", userID),
		)
	}
	
	return telegramID, nil
}

func (s *userStorage) Create(ctx context.Context, user *models.User) error {
	tx := s.db.WithContext(ctx).Create(user)
	if tx.Error != nil {
		return apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to create user")
	}
	return nil
}

func (s *userStorage) GetWithTelegramID(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	
	err := s.db.WithContext(ctx).
		Where("telegram_id IS NOT NULL AND telegram_id != ''").
		Find(&users).Error
	
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to get users with telegram ID")
	}
	
	return users, nil
}

func (s *userStorage) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, "id = ?", id).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.New(
			apperror.CodeNotFound,
			fmt.Sprintf("user %s not found", id),
		)
	}
	
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to retrieve user by ID")
	}
	
	return &user, nil
}

func (s *userStorage) GetByTelegramID(ctx context.Context, telegramID string) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, "telegram_id = ?", telegramID).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.New(
			apperror.CodeNotFound,
			fmt.Sprintf("user with telegram ID %s not found", telegramID),
		)
	}
	
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to retrieve user by telegram ID")
	}
	
	return &user, nil
}

func (s *userStorage) UpdateTelegramID(ctx context.Context, userID string, telegramID int64) error {
	result := s.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("telegram_id", telegramID)
	
	if result.Error != nil {
		return apperror.Wrap(result.Error, apperror.CodeInternal, "failed to update telegram ID")
	}
	
	if result.RowsAffected == 0 {
		return apperror.New(apperror.CodeNotFound, "user not found")
	}
	
	return nil
}

func (s *userStorage) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	
	err := s.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.CodeNotFound, "user not found")
		}
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to get user by email")
	}
	
	return &user, nil
}

func (s *userStorage) Update(ctx context.Context, user *models.User) error {
	tx := s.db.WithContext(ctx).Save(user)
	if tx.Error != nil {
		return apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to update user")
	}
	return nil
}

func (s *userStorage) Delete(ctx context.Context, id string) error {
	tx := s.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id)
	
	if tx.Error != nil {
		return apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to delete user")
	}
	
	if tx.RowsAffected == 0 {
		return apperror.New(
			apperror.CodeNotFound,
			fmt.Sprintf("user %s not found", id),
		)
	}
	
	return nil
}