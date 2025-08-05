// Package storage provides persistence implementations for domain models.
package storage

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"caviar/internal/models"
	"caviar/pkg/apperror"
)

type productStorage struct {
    db *gorm.DB
}

func NewProductStorage(db *gorm.DB) *productStorage {
    return &productStorage{
        db: db.Session(&gorm.Session{
            PrepareStmt: true,
        }),
    }
}

func (s *productStorage) Create(ctx context.Context, p *models.Product) error {
    tx := s.db.WithContext(ctx).Create(p)
    if tx.Error != nil {
        return apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to create product")
    }
    return nil
}

func (s *productStorage) GetByID(ctx context.Context, id string) (*models.Product, error) {
    var p models.Product
    err := s.db.
        WithContext(ctx).
        Preload("Variants").
        First(&p, id).
        Error

    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperror.New(
            apperror.CodeNotFound,
            fmt.Sprintf("product %s not found", id),
        )
    }
    if err != nil {
        return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to retrieve product by ID")
    }
    return &p, nil
}

func (s *productStorage) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
    var p models.Product
    err := s.db.
        WithContext(ctx).
        Preload("Variants").
        Where("slug = ?", slug).
        First(&p).
        Error

    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperror.New(
            apperror.CodeNotFound,
            fmt.Sprintf("product with slug %q not found", slug),
        )
    }
    if err != nil {
        return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to retrieve product by slug")
    }
    return &p, nil
}

func (s *productStorage) List(ctx context.Context, limit, offset int) ([]*models.Product, error) {
    var products []*models.Product
    tx := s.db.
        WithContext(ctx).
        Preload("Variants").
        Order("created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&products)

    if tx.Error != nil {
        return nil, apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to list products")
    }
    return products, nil
}

func (s *productStorage) Update(ctx context.Context, p *models.Product) error {
    tx := s.db.
        WithContext(ctx).
        Save(p)

    if tx.Error != nil {
        return apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to update product")
    }
    if tx.RowsAffected == 0 {
        return apperror.New(
            apperror.CodeNotFound,
            fmt.Sprintf("product %s not found", p.ID),
        )
    }
    return nil
}

func (s *productStorage) Delete(ctx context.Context, id string) error {
    tx := s.db.
        WithContext(ctx).
        Delete(&models.Product{}, id)

    if tx.Error != nil {
        return apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to delete product")
    }
    if tx.RowsAffected == 0 {
        return apperror.New(
            apperror.CodeNotFound,
            fmt.Sprintf("product %s not found", id),
        )
    }
    return nil
}
