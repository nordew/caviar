package dto

type ProductCreateDTO struct {
    Slug        string `json:"slug"`
    Name        string `json:"name"`
    Subtitle    string `json:"subtitle"`
    Description string `json:"description"`
    Images      []string `json:"images"`
    Variants    []VariantCreateDTO `json:"variants"`
    Details     CaviarDetailsDTO `json:"details"`
}

type VariantCreateDTO struct {
    Mass   int `json:"mass"`
    Stock  int `json:"stock"`
    Prices map[string]MoneyDTO `json:"prices"`
}

type MoneyDTO struct {
    Amount   float64 `json:"amount"`
    Currency string `json:"currency"`
}

type CaviarDetailsDTO struct {
    FishAge   string `json:"fish_age"`
    GrainSize string `json:"grain_size"`
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