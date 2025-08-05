package dto

type ProductCreateDTO struct {
    Slug        string `json:"slug"`
    Name        string `json:"name"`
    Subtitle    string `json:"subtitle"`
    Description string `json:"description"`
    Variants    []VariantCreateDTO `json:"variants"`
    Details     CaviarDetailsDTO `json:"details"`
}

type VariantCreateDTO struct {
    Mass   int `json:"mass"`
    Stock  int `json:"stock"`
    Prices map[string]MoneyDTO `json:"prices"`
}

type MoneyDTO struct {
    Amount   int `json:"amount"`
    Currency string `json:"currency"`
}

type CaviarDetailsDTO struct {
    FishAge   string `json:"fishAge"`
    GrainSize string `json:"grainSize"`
    Color     string `json:"color"`
    Taste     string `json:"taste"`
    Texture   string `json:"texture"`
    ShelfLife ShelfLifeDTO `json:"shelfLife"`
}

type ShelfLifeDTO struct {
    Duration  string `json:"duration"`
    TempRange TemperatureRangeDTO `json:"tempRange"`
}

type TemperatureRangeDTO struct {
    MinC float64 `json:"minC"`
    MaxC float64 `json:"maxC"`
}

type ProductUpdateDTO struct {
    ID          string `json:"id"`
    Slug        string `json:"slug,omitempty"`
    Name        string `json:"name,omitempty"`
    Subtitle    string `json:"subtitle,omitempty"`
    Description string `json:"description,omitempty"`
    Images      []string `json:"images,omitempty"`
    Variants    []VariantUpdateDTO `json:"variants,omitempty"`
    Details     *CaviarDetailsDTO `json:"details,omitempty"`
}

type VariantUpdateDTO struct {
    ID     string `json:"id,omitempty"`
    Mass   int    `json:"mass,omitempty"`
    Stock  int    `json:"stock,omitempty"`
    Prices map[string]MoneyDTO `json:"prices,omitempty"`
}