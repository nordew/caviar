package converter

import (
	"time"

	"caviar/internal/dto"
	"caviar/internal/models"
)

type ProductConverter struct{}

func NewProductConverter() *ProductConverter {
	return &ProductConverter{}
}

// ToResponseDTO converts a model Product to ProductResponseDTO
func (c *ProductConverter) ToResponseDTO(product *models.Product) dto.ProductResponseDTO {
	if product == nil {
		return dto.ProductResponseDTO{}
	}

	return dto.ProductResponseDTO{
		ID:          product.ID,
		Slug:        product.Slug,
		Name:        product.Name,
		Subtitle:    product.Subtitle,
		Description: product.Description,
		Variants:    c.toVariantResponseDTOs(product.Variants),
		Details:     c.toCaviarDetailsDTO(product.Details),
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}
}

// ToResponseDTOs converts a slice of model Products to ProductResponseDTOs
func (c *ProductConverter) ToResponseDTOs(products []*models.Product) []dto.ProductResponseDTO {
	if len(products) == 0 {
		return []dto.ProductResponseDTO{}
	}

	result := make([]dto.ProductResponseDTO, 0, len(products))
	for _, product := range products {
		result = append(result, c.ToResponseDTO(product))
	}
	return result
}

// toVariantResponseDTOs converts model Variants to VariantResponseDTOs
func (c *ProductConverter) toVariantResponseDTOs(variants []models.Variant) []dto.VariantResponseDTO {
	if len(variants) == 0 {
		return []dto.VariantResponseDTO{}
	}

	result := make([]dto.VariantResponseDTO, 0, len(variants))
	for _, variant := range variants {
		result = append(result, c.toVariantResponseDTO(variant))
	}
	return result
}

// toVariantResponseDTO converts a model Variant to VariantResponseDTO
func (c *ProductConverter) toVariantResponseDTO(variant models.Variant) dto.VariantResponseDTO {
	return dto.VariantResponseDTO{
		ID:        variant.ID,
		Mass:      variant.Mass,
		Stock:     variant.Stock,
		Prices:    c.toMoneyDTOMap(variant.Prices),
		CreatedAt: variant.CreatedAt.Format(time.RFC3339),
		UpdatedAt: variant.UpdatedAt.Format(time.RFC3339),
	}
}

// toMoneyDTOMap converts model MoneyMap to map[string]MoneyDTO
func (c *ProductConverter) toMoneyDTOMap(prices models.MoneyMap) map[string]dto.MoneyDTO {
	if len(prices) == 0 {
		return map[string]dto.MoneyDTO{}
	}

	result := make(map[string]dto.MoneyDTO, len(prices))
	for region, money := range prices {
		result[region] = dto.MoneyDTO{
			Amount:   money.Amount,
			Currency: money.Currency,
		}
	}
	return result
}

// toCaviarDetailsDTO converts model CaviarDetails to CaviarDetailsDTO
func (c *ProductConverter) toCaviarDetailsDTO(details models.CaviarDetails) dto.CaviarDetailsDTO {
	return dto.CaviarDetailsDTO{
		FishAge:   details.FishAge,
		GrainSize: details.GrainSize,
		Color:     details.Color,
		Taste:     details.Taste,
		Texture:   details.Texture,
		ShelfLife: dto.ShelfLifeDTO{
			Duration: details.ShelfLife.Duration,
			TempRange: dto.TemperatureRangeDTO{
				MinC: details.ShelfLife.TempRange.MinC,
				MaxC: details.ShelfLife.TempRange.MaxC,
			},
		},
	}
}

// FromCreateDTO converts ProductCreateDTO to model Product
func (c *ProductConverter) FromCreateDTO(input dto.ProductCreateDTO) (*models.Product, error) {
	return models.NewProduct(input)
}

// ToUpdateDTO prepares a ProductUpdateDTO from existing product (helper for updates)
func (c *ProductConverter) ToUpdateDTO(product *models.Product) dto.ProductUpdateDTO {
	if product == nil {
		return dto.ProductUpdateDTO{}
	}

	updateDTO := dto.ProductUpdateDTO{
		ID:          product.ID,
		Slug:        product.Slug,
		Name:        product.Name,
		Subtitle:    product.Subtitle,
		Description: product.Description,
	}

	// Convert variants
	if len(product.Variants) > 0 {
		updateDTO.Variants = make([]dto.VariantUpdateDTO, 0, len(product.Variants))
		for _, variant := range product.Variants {
			updateDTO.Variants = append(updateDTO.Variants, dto.VariantUpdateDTO{
				ID:     variant.ID,
				Mass:   variant.Mass,
				Stock:  variant.Stock,
				Prices: c.toMoneyDTOMap(variant.Prices),
			})
		}
	}

	// Convert details
	detailsDTO := c.toCaviarDetailsDTO(product.Details)
	updateDTO.Details = &detailsDTO

	return updateDTO
}