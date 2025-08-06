package converter

import (
	"time"

	"caviar/internal/dto"
	"caviar/internal/models"
)

type OrderConverter struct{}

func NewOrderConverter() *OrderConverter {
	return &OrderConverter{}
}

// ToResponseDTO converts a model Order to OrderResponseDTO
func (c *OrderConverter) ToResponseDTO(order *models.Order) dto.OrderResponseDTO {
	if order == nil {
		return dto.OrderResponseDTO{}
	}

	return dto.OrderResponseDTO{
		ID:          order.ID,
		OrderNumber: order.OrderNumber,
		CustomerInfo: c.toCustomerInfoDTO(order.CustomerInfo),
		DeliveryInfo: c.toDeliveryInfoDTO(order.DeliveryInfo),
		Items:        c.toOrderItemsDTO(order.Items),
		TotalAmount:  c.toMoneyDTO(order.TotalAmount),
		Status:       string(order.Status),
		Notes:        order.Notes,
		CreatedAt:    order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    order.UpdatedAt.Format(time.RFC3339),
	}
}

// ToResponseDTOs converts a slice of model Orders to OrderResponseDTOs
func (c *OrderConverter) ToResponseDTOs(orders []*models.Order) []dto.OrderResponseDTO {
	if len(orders) == 0 {
		return []dto.OrderResponseDTO{}
	}

	result := make([]dto.OrderResponseDTO, 0, len(orders))
	for _, order := range orders {
		result = append(result, c.ToResponseDTO(order))
	}
	return result
}

// ToListResponseDTO converts orders with pagination info to OrderListResponseDTO
func (c *OrderConverter) ToListResponseDTO(orders []*models.Order, page, limit int, total int64) dto.OrderListResponseDTO {
	return dto.OrderListResponseDTO{
		Orders: c.ToResponseDTOs(orders),
		Total:  int(total),
		Page:   page,
		Limit:  limit,
	}
}

// toCustomerInfoDTO converts model CustomerInfo to CustomerInfoDTO
func (c *OrderConverter) toCustomerInfoDTO(info models.CustomerInfo) dto.CustomerInfoDTO {
	return dto.CustomerInfoDTO{
		FirstName: info.FirstName,
		LastName:  info.LastName,
		FullName:  info.FullName,
		Phone:     info.Phone,
		Email:     info.Email,
	}
}

// toDeliveryInfoDTO converts model DeliveryInfo to DeliveryInfoDTO
func (c *OrderConverter) toDeliveryInfoDTO(info models.DeliveryInfo) dto.DeliveryInfoDTO {
	return dto.DeliveryInfoDTO{
		Type:         string(info.Type),
		Country:      info.Country,
		City:         info.City,
		Address:      info.Address,
		PostOffice:   info.PostOffice,
		Instructions: info.Instructions,
	}
}

// toOrderItemsDTO converts model OrderItems to OrderItemResponseDTOs
func (c *OrderConverter) toOrderItemsDTO(items []models.OrderItem) []dto.OrderItemResponseDTO {
	if len(items) == 0 {
		return []dto.OrderItemResponseDTO{}
	}

	result := make([]dto.OrderItemResponseDTO, 0, len(items))
	productConverter := NewProductConverter()
	
	for _, item := range items {
		itemResponse := dto.OrderItemResponseDTO{
			ID:         item.ID,
			ProductID:  item.ProductID,
			VariantID:  item.VariantID,
			Quantity:   item.Quantity,
			UnitPrice:  c.toMoneyDTO(item.UnitPrice),
			TotalPrice: c.toMoneyDTO(item.TotalPrice),
		}

		if item.Product != nil {
			productDTO := productConverter.ToResponseDTO(item.Product)
			itemResponse.Product = &productDTO
		}

		result = append(result, itemResponse)
	}
	return result
}

// toMoneyDTO converts model Money to MoneyDTO
func (c *OrderConverter) toMoneyDTO(money models.Money) dto.MoneyDTO {
	return dto.MoneyDTO{
		Amount:   money.Amount,
		Currency: money.Currency,
	}
}

// FromCreateDTO converts OrderCreateDTO to model Order (for use in services if needed)
func (c *OrderConverter) FromCreateDTO(input dto.OrderCreateDTO) (*models.Order, error) {
	return models.NewOrder(input)
}