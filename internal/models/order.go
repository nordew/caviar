package models

import (
	"fmt"
	"time"

	"caviar/internal/dto"
	"caviar/pkg/apperror"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type DeliveryType string

const (
	DeliveryTypePostOffice DeliveryType = "post_office"
	DeliveryTypeCourier    DeliveryType = "courier"
	DeliveryTypeAddress    DeliveryType = "address"
)

type Order struct {
	ID           string      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderNumber  string      `gorm:"uniqueIndex;not null"`
	CustomerInfo CustomerInfo `gorm:"type:jsonb;not null"`
	DeliveryInfo DeliveryInfo `gorm:"type:jsonb;not null"`
	Items        []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	TotalAmount  Money       `gorm:"type:jsonb;not null"`
	Status       OrderStatus `gorm:"type:varchar(20);not null;default:'pending'"`
	Notes        string      `gorm:"type:text"`
	CreatedAt    time.Time   `gorm:"not null;default:now()"`
	UpdatedAt    time.Time   `gorm:"not null;default:now()"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID   string    `gorm:"type:uuid;not null"`
	ProductID string    `gorm:"type:uuid;not null"`
	VariantID string    `gorm:"type:uuid;not null"`
	Quantity  int       `gorm:"not null"`
	UnitPrice Money     `gorm:"type:jsonb;not null"`
	TotalPrice Money    `gorm:"type:jsonb;not null"`
	Product   *Product  `gorm:"foreignKey:ProductID"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

type CustomerInfo struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	FullName  string `json:"full_name,omitempty"`
	Phone     string `json:"phone"`
	Email     string `json:"email,omitempty"`
}

type DeliveryInfo struct {
	Type         DeliveryType `json:"type"`
	Country      string       `json:"country"`
	City         string       `json:"city"`
	Address      string       `json:"address,omitempty"`
	PostOffice   string       `json:"post_office,omitempty"`
	Instructions string       `json:"instructions,omitempty"`
}

func NewOrder(input dto.OrderCreateDTO) (*Order, error) {
	now := time.Now()

	if len(input.Items) == 0 {
		return nil, apperror.New(apperror.CodeInvalidInput, "order must contain at least one item")
	}

	customerInfo, err := validateCustomerInfo(input.CustomerInfo, input.DeliveryInfo.Country)
	if err != nil {
		return nil, err
	}

	deliveryInfo, err := validateDeliveryInfo(input.DeliveryInfo)
	if err != nil {
		return nil, err
	}

	var orderItems []OrderItem
	var totalAmount int
	currency := ""

	for i, item := range input.Items {
		if item.ProductID == "" {
			return nil, apperror.New(apperror.CodeInvalidInput, fmt.Sprintf("product ID is required for item %d", i+1))
		}
		if item.VariantID == "" {
			return nil, apperror.New(apperror.CodeInvalidInput, fmt.Sprintf("variant ID is required for item %d", i+1))
		}
		if item.Quantity <= 0 {
			return nil, apperror.New(apperror.CodeInvalidInput, fmt.Sprintf("quantity must be greater than 0 for item %d", i+1))
		}
		if item.UnitPrice.Amount <= 0 {
			return nil, apperror.New(apperror.CodeInvalidInput, fmt.Sprintf("unit price must be greater than 0 for item %d", i+1))
		}
		if item.UnitPrice.Currency == "" {
			return nil, apperror.New(apperror.CodeInvalidInput, fmt.Sprintf("currency is required for item %d", i+1))
		}

		if currency == "" {
			currency = item.UnitPrice.Currency
		} else if currency != item.UnitPrice.Currency {
			return nil, apperror.New(apperror.CodeInvalidInput, "all items must have the same currency")
		}

		itemTotal := item.UnitPrice.Amount * item.Quantity
		totalAmount += itemTotal

		orderItems = append(orderItems, OrderItem{
			ID:         uuid.New().String(),
			ProductID:  item.ProductID,
			VariantID:  item.VariantID,
			Quantity:   item.Quantity,
			UnitPrice:  Money{
				Amount:   item.UnitPrice.Amount,
				Currency: item.UnitPrice.Currency,
			},
			TotalPrice: Money{
				Amount:   itemTotal,
				Currency: currency,
			},
			CreatedAt:  now,
		})
	}

	orderNumber := generateOrderNumber()

	order := &Order{
		ID:           uuid.New().String(),
		OrderNumber:  orderNumber,
		CustomerInfo: *customerInfo,
		DeliveryInfo: *deliveryInfo,
		Items:        orderItems,
		TotalAmount:  Money{
			Amount:   totalAmount,
			Currency: currency,
		},
		Status:       OrderStatusPending,
		Notes:        input.Notes,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return order, nil
}

func validateCustomerInfo(info dto.CustomerInfoDTO, country string) (*CustomerInfo, error) {
	customerInfo := &CustomerInfo{
		Phone: info.Phone,
		Email: info.Email,
	}

	if info.Phone == "" {
		return nil, apperror.New(apperror.CodeInvalidInput, "phone number is required")
	}

	// Accept either full name or first/last name format globally
	if info.FullName != "" {
		customerInfo.FullName = info.FullName
	} else if info.FirstName != "" && info.LastName != "" {
		customerInfo.FirstName = info.FirstName
		customerInfo.LastName = info.LastName
	} else {
		return nil, apperror.New(apperror.CodeInvalidInput, "either full name or both first and last name are required")
	}

	return customerInfo, nil
}

func validateDeliveryInfo(info dto.DeliveryInfoDTO) (*DeliveryInfo, error) {
	deliveryInfo := &DeliveryInfo{
		Type:         DeliveryType(info.Type),
		Country:      info.Country,
		City:         info.City,
		Address:      info.Address,
		PostOffice:   info.PostOffice,
		Instructions: info.Instructions,
	}

	if info.Country == "" {
		return nil, apperror.New(apperror.CodeInvalidInput, "country is required")
	}
	if info.City == "" {
		return nil, apperror.New(apperror.CodeInvalidInput, "city is required")
	}

	switch DeliveryType(info.Type) {
	case DeliveryTypePostOffice:
		if info.PostOffice == "" {
			return nil, apperror.New(apperror.CodeInvalidInput, "post office is required for post office delivery")
		}
	case DeliveryTypeCourier:
		if info.Address == "" {
			return nil, apperror.New(apperror.CodeInvalidInput, "address is required for courier delivery")
		}
	case DeliveryTypeAddress:
		if info.Address == "" {
			return nil, apperror.New(apperror.CodeInvalidInput, "address is required for address delivery")
		}
	default:
		return nil, apperror.New(apperror.CodeInvalidInput, "invalid delivery type")
	}

	return deliveryInfo, nil
}

func generateOrderNumber() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("ORD%s-%d", uuid.New().String()[:8], timestamp%10000)
}

func (o *Order) UpdateStatus(status OrderStatus) error {
	switch status {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusProcessing, OrderStatusShipped, OrderStatusDelivered, OrderStatusCancelled:
		o.Status = status
		o.UpdatedAt = time.Now()
		return nil
	default:
		return apperror.New(apperror.CodeInvalidInput, "invalid order status")
	}
}