package service

import (
	"caviar/internal/dto"
	"caviar/internal/models"
	"caviar/internal/types"
	"context"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type productService struct {
	productStorage ProductStorage
	minioClient *minio.Client
	logger *zap.Logger
}

func NewProductService(
	productStorage ProductStorage,
	minioClient *minio.Client,
	logger *zap.Logger,	
) *productService {
	return &productService{
		productStorage: productStorage,
		minioClient: minioClient,
		logger: logger,
	}
}


func (s *productService) Create(ctx context.Context, input *dto.ProductCreateDTO) error {
	product, err := models.NewProduct(*input)
	if err != nil {
		s.logger.Error("failed to create product", zap.Error(err))
		return err
	}

	if err := s.productStorage.Create(ctx, product); err != nil {
		s.logger.Error("failed to create product", zap.Error(err))
		return err
	}

	return nil
}

func (s *productService) List(
	ctx context.Context, 
	isAuthenticated bool,
	filter *types.ProductFilter,
) ([]*models.Product, error) {
	// Only authenticated users can see all products (including inactive ones)
	if !isAuthenticated {
		filter.ShowAll = false
	}

	return s.productStorage.List(ctx, filter)
}

func (s *productService) Update(ctx context.Context, input *dto.ProductUpdateDTO) error {
	if err := s.productStorage.Update(ctx, input); err != nil {
		s.logger.Error("failed to update product", zap.Error(err))
		return err
	}

	return nil
}

func (s *productService) Delete(ctx context.Context, id string) error {
	if err := s.productStorage.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete product", zap.Error(err))
		return err
	}

	return nil
}