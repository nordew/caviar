package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"caviar/internal/dto"
	"caviar/internal/models"
	"caviar/internal/types"
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
    // Set ProductID for all variants
    for i := range p.Variants {
        p.Variants[i].ProductID = p.ID
    }
    
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
        Where("id = ?", id).
        First(&p).
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

func (s *productStorage) List(ctx context.Context, filter *types.ProductFilter) ([]*models.Product, error) {
    var products []*models.Product
    
    if filter == nil {
        filter = types.DefaultProductFilter()
    }
    
    if err := filter.Validate(); err != nil {
        return nil, apperror.New(apperror.CodeInvalidInput, "invalid filter parameters")
    }
    
    tx := s.db.WithContext(ctx)
    
    // Always apply filters, even if filter.IsEmpty() returns true
    tx = s.applyProductFilters(tx, filter)
    
    if filter.IncludeVariants {
        tx = tx.Preload("Variants")
    }
    
    sortField := s.getSortField(filter.SortBy)
    sortOrder := filter.SortOrder
    tx = tx.Order(fmt.Sprintf("%s %s", sortField, strings.ToUpper(sortOrder)))
    
    tx = tx.Limit(filter.Limit).Offset(filter.Offset)
    
    if err := tx.Find(&products).Error; err != nil {
        return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to list products")
    }
    
    return products, nil
}

func (s *productStorage) applyProductFilters(tx *gorm.DB, filter *types.ProductFilter) *gorm.DB {
    var zeroTime time.Time
    
    if filter.Slug != "" {
        tx = tx.Where("slug = ?", filter.Slug)
    }
    if filter.Name != "" {
        tx = tx.Where("name ILIKE ?", "%"+filter.Name+"%")
    }
    if filter.Subtitle != "" {
        tx = tx.Where("subtitle ILIKE ?", "%"+filter.Subtitle+"%")
    }
    if filter.Description != "" {
        tx = tx.Where("description ILIKE ?", "%"+filter.Description+"%")
    }
	if !filter.ShowAll {
		tx = tx.Where("is_active = ?", true)
	}
    
    if !filter.CreatedAfter.Equal(zeroTime) {
        tx = tx.Where("created_at >= ?", filter.CreatedAfter)
    }
    if !filter.CreatedBefore.Equal(zeroTime) {
        tx = tx.Where("created_at <= ?", filter.CreatedBefore)
    }
    if !filter.UpdatedAfter.Equal(zeroTime) {
        tx = tx.Where("updated_at >= ?", filter.UpdatedAfter)
    }
    if !filter.UpdatedBefore.Equal(zeroTime) {
        tx = tx.Where("updated_at <= ?", filter.UpdatedBefore)
    }
    
    if filter.Search != "" {
        searchTerm := "%" + filter.Search + "%"
        tx = tx.Where("name ILIKE ? OR subtitle ILIKE ? OR description ILIKE ?", 
            searchTerm, searchTerm, searchTerm)
    }
    
    return tx
}

func (s *productStorage) getSortField(sortBy string) string {
    switch sortBy {
    case "name":
        return "name"
    case "slug":
        return "slug"
    case "createdAt":
        return "created_at"
    case "updatedAt":
        return "updated_at"
    default:
        return "created_at"
    }
}

func (s *productStorage) Update(ctx context.Context, input *dto.ProductUpdateDTO) error {
    existingProduct, err := s.GetByID(ctx, input.ID)
    if err != nil {
        return err
    }
    
    if input.Slug != "" {
        existingProduct.Slug = input.Slug
    }
    if input.Name != "" {
        existingProduct.Name = input.Name
    }
    if input.Subtitle != "" {
        existingProduct.Subtitle = input.Subtitle
    }
    if input.Description != "" {
        existingProduct.Description = input.Description
    }
    
    if input.Details != nil {
        existingProduct.Details = models.CaviarDetails{
            FishAge:   input.Details.FishAge,
            GrainSize: input.Details.GrainSize,
            Color:     input.Details.Color,
            Taste:     input.Details.Taste,
            Texture:   input.Details.Texture,
            ShelfLife: models.ShelfLife{
                Duration: input.Details.ShelfLife.Duration,
                TempRange: models.TemperatureRange{
                    MinC: input.Details.ShelfLife.TempRange.MinC,
                    MaxC: input.Details.ShelfLife.TempRange.MaxC,
                },
            },
        }
    }
    
    if input.Variants != nil {
        var variants []models.Variant
        for _, v := range input.Variants {
            variant := models.Variant{
                ID:        v.ID,
                ProductID: input.ID,
                Mass:      v.Mass,
                Stock:     v.Stock,
                Prices:    make(models.MoneyMap),
            }
            
            for region, price := range v.Prices {
                variant.Prices[region] = models.Money{
                    Amount:   price.Amount,
                    Currency: price.Currency,
                }
            }
            
            variants = append(variants, variant)
        }
        existingProduct.Variants = variants
    }
    
    existingProduct.UpdatedAt = time.Now().UTC()
    
    tx := s.db.WithContext(ctx).Save(existingProduct)
    if tx.Error != nil {
        return apperror.Wrap(tx.Error, apperror.CodeInternal, "failed to update product")
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

func (s *productStorage) GetVariantByID(ctx context.Context, productID, variantID string) (*models.Variant, error) {
    var variant models.Variant
    err := s.db.WithContext(ctx).
        Where("id = ? AND product_id = ?", variantID, productID).
        First(&variant).Error

    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperror.New(apperror.CodeNotFound, "variant not found")
    }
    if err != nil {
        return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to get variant")
    }

    return &variant, nil
}

func (s *productStorage) UpdateVariantStock(ctx context.Context, variantID string, stockChange int) error {
    result := s.db.WithContext(ctx).
        Model(&models.Variant{}).
        Where("id = ?", variantID).
        Update("stock", gorm.Expr("stock + ?", stockChange))

    if result.Error != nil {
        return apperror.Wrap(result.Error, apperror.CodeInternal, "failed to update variant stock")
    }

    if result.RowsAffected == 0 {
        return apperror.New(apperror.CodeNotFound, "variant not found")
    }

    return nil
}
