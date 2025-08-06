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
	GetVariantByID(ctx context.Context, productID, variantID string) (*models.Variant, error)
	List(ctx context.Context, filter *types.ProductFilter) ([]*models.Product, error)
	Update(ctx context.Context, input *dto.ProductUpdateDTO) error
	UpdateVariantStock(ctx context.Context, variantID string, stockChange int) error
	Delete(ctx context.Context, id string) error
}

type OrderStorage interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error)
	List(ctx context.Context, filter *types.OrderFilter) ([]*models.Order, int64, error)
	Update(ctx context.Context, order *models.Order) error
	UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error
	Delete(ctx context.Context, id string) error
	GetOrderStatistics(ctx context.Context) (map[string]any, error)
}
