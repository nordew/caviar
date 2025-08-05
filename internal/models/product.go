package models

import (
	"time"

	"caviar/internal/dto"
	"caviar/pkg/apperror"
)

type Product struct {
    ID          string        `db:"id"`          
    Slug        string        `db:"slug"`        
    Name        string        `db:"name"`        
    Subtitle    string        `db:"subtitle"`    
    Description string        `db:"description"`
    Images      []string      `db:"images"`      
    Variants    []Variant                     
    Details     CaviarDetails `db:"details"`     
    CreatedAt   time.Time     `db:"created_at"`
    UpdatedAt   time.Time     `db:"updated_at"`
}

type Variant struct {
    ID        string            `db:"id"`
    ProductID string            `db:"product_id"`
    Mass      int              `db:"mass"`
    Stock     int              `db:"stock"`
    Prices    map[string]Money
}

type Money struct {
    Amount   float64 
    Currency string 
}

type CaviarDetails struct {
    FishAge   string        `db:"fish_age"`   
    GrainSize string        `db:"grain_size"` 
    Color     string        `db:"color"`      
    Taste     string        `db:"taste"`      
    Texture   string        `db:"texture"`    
    ShelfLife ShelfLife     `db:"shelf_life"`
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
    // required string fields
    if input.Slug == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "slug is required")
    }
    if input.Name == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "name is required")
    }
    if input.Subtitle == "" {
        return nil, apperror.New(apperror.CodeInvalidInput, "subtitle is required")
    }
    // at least one image
    if len(input.Images) == 0 {
        return nil, apperror.New(apperror.CodeInvalidInput, "at least one image is required")
    }
    // variants
    if len(input.Variants) == 0 {
        return nil, apperror.New(apperror.CodeInvalidInput, "at least one variant is required")
    }
    var variants []Variant
    for i, v := range input.Variants {
        if v.Mass <= 0 {
            return nil, apperror.New(apperror.CodeInvalidInput, "variant mass must be > 0 (index " + string(i) + ")")
        }
        if v.Stock < 0 {
            return nil, apperror.New(apperror.CodeInvalidInput, "variant stock cannot be negative (index " + string(i) + ")")
        }
        if len(v.Prices) == 0 {
            return nil, apperror.New(apperror.CodeInvalidInput, "variant must have at least one price (index " + string(i) + ")")
        }
        m := make(map[string]Money, len(v.Prices))
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
            Mass:   v.Mass,
            Stock:  v.Stock,
            Prices: m,
        })
    }
    // details
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
    // temperature range: no further checks here, but you could enforce MinC < MaxC

    // assemble Product
    now := time.Now().UTC()
    prod := &Product{
        Slug:        input.Slug,
        Name:        input.Name,
        Subtitle:    input.Subtitle,
        Description: input.Description,
        Images:      input.Images,
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
        CreatedAt: now,
        UpdatedAt: now,
    }

    return prod, nil
}