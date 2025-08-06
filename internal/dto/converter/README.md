# DTO Converter Package

This package provides conversion utilities between domain models and Data Transfer Objects (DTOs).

## Purpose

The converter package centralizes all conversion logic between:
- Domain models (`internal/models`) - Business entities
- DTOs (`internal/dto`) - API request/response structures

## Architecture

### Main Converter
- `Converter` - Main converter struct that aggregates all specific converters
- `Default()` - Returns singleton converter instance for convenience

### Order Converter
- `ToResponseDTO()` - Converts single Order model to OrderResponseDTO
- `ToResponseDTOs()` - Converts slice of Orders to OrderResponseDTOs
- `ToListResponseDTO()` - Converts Orders with pagination to OrderListResponseDTO
- `FromCreateDTO()` - Converts OrderCreateDTO to Order model

### Product Converter
- `ToResponseDTO()` - Converts single Product model to ProductResponseDTO
- `ToResponseDTOs()` - Converts slice of Products to ProductResponseDTOs
- `FromCreateDTO()` - Converts ProductCreateDTO to Product model
- `ToUpdateDTO()` - Prepares ProductUpdateDTO from existing Product

## Usage

### In Handlers

```go
// Initialize converter in handler
type Handler struct {
    converter *converter.Converter
    // ... other fields
}

func NewHandler() *Handler {
    return &Handler{
        converter: converter.NewConverter(),
        // ... other initialization
    }
}

// Use converter in endpoints
func (h *Handler) getOrder(c *gin.Context) {
    order, err := h.orderService.GetByID(ctx, id)
    if err != nil {
        h.handleError(c, err)
        return
    }
    
    response := h.converter.Order.ToResponseDTO(order)
    h.handleSuccess(c, http.StatusOK, response)
}
```

### Standalone Usage

```go
// Using singleton instance
converter := converter.Default()
dto := converter.Product.ToResponseDTO(product)

// Creating new instance
conv := converter.NewConverter()
dto := conv.Order.ToResponseDTO(order)
```

## Benefits

1. **Separation of Concerns** - Conversion logic is isolated from business logic and handlers
2. **Reusability** - Same converters can be used across different handlers
3. **Testability** - Converters can be unit tested independently
4. **Maintainability** - All conversion logic in one place
5. **Type Safety** - Strongly typed conversions with compile-time checks

## Adding New Converters

To add a new converter:

1. Create a new file (e.g., `user.go`)
2. Define the converter struct and methods
3. Add it to the main `Converter` struct
4. Initialize it in `NewConverter()`

Example:
```go
// user.go
type UserConverter struct{}

func NewUserConverter() *UserConverter {
    return &UserConverter{}
}

func (c *UserConverter) ToResponseDTO(user *models.User) dto.UserResponseDTO {
    // conversion logic
}

// converter.go
type Converter struct {
    Order   *OrderConverter
    Product *ProductConverter
    User    *UserConverter // Add new converter
}

func NewConverter() *Converter {
    return &Converter{
        Order:   NewOrderConverter(),
        Product: NewProductConverter(),
        User:    NewUserConverter(), // Initialize
    }
}
```