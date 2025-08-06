package dto

type OrderCreateDTO struct {
	CustomerInfo CustomerInfoDTO  `json:"customerInfo" binding:"required"`
	DeliveryInfo DeliveryInfoDTO  `json:"deliveryInfo" binding:"required"`
	Items        []OrderItemDTO   `json:"items" binding:"required,min=1"`
	Notes        string           `json:"notes"`
}

type CustomerInfoDTO struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	FullName  string `json:"fullName"`
	Phone     string `json:"phone" binding:"required"`
	Email     string `json:"email"`
}

type DeliveryInfoDTO struct {
	Type         string `json:"type" binding:"required,oneof=post_office courier address"`
	Country      string `json:"country" binding:"required"`
	City         string `json:"city" binding:"required"`
	Address      string `json:"address"`
	PostOffice   string `json:"postOffice"`
	Instructions string `json:"instructions"`
}

type OrderItemDTO struct {
	ProductID string    `json:"productId" binding:"required"`
	VariantID string    `json:"variantId" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
	UnitPrice MoneyDTO  `json:"unitPrice" binding:"required"`
}


type OrderResponseDTO struct {
	ID           string               `json:"id"`
	OrderNumber  string               `json:"orderNumber"`
	CustomerInfo CustomerInfoDTO      `json:"customerInfo"`
	DeliveryInfo DeliveryInfoDTO      `json:"deliveryInfo"`
	Items        []OrderItemResponseDTO `json:"items"`
	TotalAmount  MoneyDTO             `json:"totalAmount"`
	Status       string               `json:"status"`
	Notes        string               `json:"notes"`
	CreatedAt    string               `json:"createdAt"`
	UpdatedAt    string               `json:"updatedAt"`
}

type OrderItemResponseDTO struct {
	ID         string      `json:"id"`
	ProductID  string      `json:"productId"`
	VariantID  string      `json:"variantId"`
	Quantity   int         `json:"quantity"`
	UnitPrice  MoneyDTO    `json:"unitPrice"`
	TotalPrice MoneyDTO    `json:"totalPrice"`
	Product    *ProductResponseDTO `json:"product,omitempty"`
}

type OrderStatusUpdateDTO struct {
	Status string `json:"status" binding:"required,oneof=pending confirmed processing shipped delivered cancelled"`
}

type OrderListResponseDTO struct {
	Orders []OrderResponseDTO `json:"orders"`
	Total  int                `json:"total"`
	Page   int                `json:"page"`
	Limit  int                `json:"limit"`
}

// Regional order DTO that adapts based on country
type RegionalOrderDTO struct {
	CustomerInfo RegionalCustomerInfoDTO `json:"customerInfo" binding:"required"`
	DeliveryInfo RegionalDeliveryInfoDTO `json:"deliveryInfo" binding:"required"`
	Items        []OrderItemDTO          `json:"items" binding:"required,min=1"`
	Notes        string                  `json:"notes"`
}

type RegionalCustomerInfoDTO struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	FullName  string `json:"fullName"`
	Phone     string `json:"phone" binding:"required"`
	Email     string `json:"email"`
}

type RegionalDeliveryInfoDTO struct {
	Type         string `json:"type" binding:"required"`
	Country      string `json:"country" binding:"required"`
	City         string `json:"city" binding:"required"`
	Address      string `json:"address"`
	PostOffice   string `json:"post_office"`
	Instructions string `json:"instructions"`
}

// ToOrderCreateDTO converts RegionalOrderDTO to OrderCreateDTO
func (r *RegionalOrderDTO) ToOrderCreateDTO() OrderCreateDTO {
	return OrderCreateDTO{
		CustomerInfo: CustomerInfoDTO{
			FirstName: r.CustomerInfo.FirstName,
			LastName:  r.CustomerInfo.LastName,
			FullName:  r.CustomerInfo.FullName,
			Phone:     r.CustomerInfo.Phone,
			Email:     r.CustomerInfo.Email,
		},
		DeliveryInfo: DeliveryInfoDTO{
			Type:         r.DeliveryInfo.Type,
			Country:      r.DeliveryInfo.Country,
			City:         r.DeliveryInfo.City,
			Address:      r.DeliveryInfo.Address,
			PostOffice:   r.DeliveryInfo.PostOffice,
			Instructions: r.DeliveryInfo.Instructions,
		},
		Items: r.Items,
		Notes: r.Notes,
	}
}
