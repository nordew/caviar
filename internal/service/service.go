package service

import (
	"caviar/internal/dto"
	"caviar/internal/models"
	"caviar/internal/types"
	"context"
)

type ProductStorage interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id string) (*models.Product, error)
	GetBySlug(ctx context.Context, slug string) (*models.Product, error)
	List(ctx context.Context, filter *types.ProductFilter) ([]*models.Product, error)
	Update(ctx context.Context, input *dto.ProductUpdateDTO) error
	Delete(ctx context.Context, id string) error
}