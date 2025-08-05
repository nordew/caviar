package models

import (
	"strconv"
	"time"

	"caviar/internal/dto"
	"caviar/pkg/apperror"

	"github.com/google/uuid"
)

type Product struct {
    ID          string        `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Slug        string        `gorm:"uniqueIndex;not null"`
    Name        string        `gorm:"not null"`
    Subtitle    string        `gorm:"not null"`
    Description string        `gorm:"type:text"`
    Variants    []Variant     `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
    Details     CaviarDetails `gorm:"type:jsonb;not null;default:'{}'::jsonb"`
    IsActive    bool          `gorm:"default:true"`
    CreatedAt   time.Time     `gorm:"not null;default:now()"`
    UpdatedAt   time.Time     `gorm:"not null;default:now()"`
}

func (Product) TableName() string {
    return "products"
}

type Variant struct {
    ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    ProductID string    `gorm:"type:uuid;not null"`
    Mass      int       `gorm:"not null"`
    Stock     int       `gorm:"not null;default:0"`
    Prices    MoneyMap  `gorm:"type:jsonb;not null;default:'{}'::jsonb"`
    CreatedAt time.Time `gorm:"not null;default:now()"`
    UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (Variant) TableName() string {
    return "product_variants"
}

type Money struct {
    Amount   int 
    Currency string 
}

type MoneyMap map[string]Money

type CaviarDetails struct {
    FishAge   string 
    GrainSize string 
    Color     string 
    Taste     string 
    Texture   string 
    ShelfLife ShelfLife 
}

type ShelfLife struct {
    Duration  string           
    TempRange TemperatureRange 
}

type TemperatureRange struct {
    MinC float64 // e.g. -2
    MaxC float64 // e.g. +2
}

func NewProduct(input dto.ProductCreateDTO) (*Product, error) {
    now := time.Now()

    if input.Slug == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "slug is required")
    }
    if input.Name == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "name is required")
    }
    if input.Subtitle == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "subtitle is required")
    }
    if len(input.Variants) == 0 {
        return nil, apperror.New(apperror.CodeInvalidInput, "at least one variant is required")
    }
    var variants []Variant
    for i, v := range input.Variants {
        if v.Mass <= 0 {
            return nil, apperror.New(apperror.CodeInvalidInput, "variant mass must be > 0 (index " + strconv.Itoa(i) + ")")
        }
        if v.Stock < 0 {
            return nil, apperror.New(apperror.CodeInvalidInput, "variant stock cannot be negative (index " + strconv.Itoa(i) + ")")
        }
        if len(v.Prices) == 0 {
            return nil, apperror.New(apperror.CodeInvalidInput, "variant must have at least one price (index " + strconv.Itoa(i) + ")")
        }
        m := make(MoneyMap, len(v.Prices))
        for region, p := range v.Prices {
            if p.Amount <= 0 {
                return nil, apperror.New(apperror.CodeInvalidInput, "price amount must be > 0 for region " + region)
            }
            if p.Currency == "" {
                return nil, apperror.New(apperror.CodeInvalidInput, "currency is required for region " + region)
            }
            m[region] = Money{
                Amount:   p.Amount,
                Currency: p.Currency,
            }
        }
        variants = append(variants, Variant{
            ID:        uuid.New().String(),
            Mass:      v.Mass,
            Stock:     v.Stock,
            Prices:    m,
            CreatedAt: now,
            UpdatedAt: now,
        })
    }
    d := input.Details
    if d.FishAge == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "fish age is required in details")
    }
    if d.GrainSize == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "grain size is required in details")
    }
    if d.Color == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "color is required in details")
    }
    if d.Taste == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "taste is required in details")
    }
    if d.Texture == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "texture is required in details")
    }
    sl := d.ShelfLife
    if sl.Duration == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "shelf life duration is required")
    }
    prod := &Product{
        ID:          uuid.New().String(),
        Slug:        input.Slug,
        Name:        input.Name,
        Subtitle:    input.Subtitle,
        Description: input.Description,
        Variants:    variants,
        Details: CaviarDetails{
            FishAge:   d.FishAge,
            GrainSize: d.GrainSize,
            Color:     d.Color,
            Taste:     d.Taste,
            Texture:   d.Texture,
            ShelfLife: ShelfLife{
                Duration: sl.Duration,
                TempRange: TemperatureRange{
                    MinC: sl.TempRange.MinC,
                    MaxC: sl.TempRange.MaxC,
                },
            },
        },
        IsActive:  false,
        CreatedAt: now,
        UpdatedAt: now,
    }

    return prod, nil
}