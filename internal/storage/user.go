// Package storage implements persistence for application entities.
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

func (s *userStorage) Create(ctx context.Context, user *models.User) error {
    tx := s.db.WithContext(ctx).Create(user)
    if tx.Error != nil {
        return apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to create user")
    }

    return nil
}

func (s *userStorage) GetByID(ctx context.Context, id string) (*models.User, error) {
    var user models.User
    err := s.db.WithContext(ctx).First(&user, "id = ?", id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperror.New(apperror.CodeNotFound, fmt.Sprintf("user %s not found", id))
    }
    if err != nil {
        return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to retrieve user")
    }

    return &user, nil
}

func (s *userStorage) Update(ctx context.Context, user *models.User) error {
    tx := s.db.WithContext(ctx).Save(user)
    if tx.Error != nil {
        return apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to update user")
    }
    if tx.RowsAffected == 0 {
        return apperror.New(apperror.CodeNotFound, fmt.Sprintf("user %s not found", user.ID))
    }

    return nil
}

func (s *userStorage) Delete(ctx context.Context, id string) error {
    tx := s.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id)
    if tx.Error != nil {
        return apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to delete user")
    }
    if tx.RowsAffected == 0 {
        return apperror.New(apperror.CodeNotFound, fmt.Sprintf("user %s not found", id))
    }
	
    return nil
}
