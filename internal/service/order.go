package service

import (
	"context"
	"fmt"

	"caviar/internal/dto"
	"caviar/internal/models"
	"caviar/internal/types"
	"caviar/pkg/apperror"

	"go.uber.org/zap"
)




type OrderService struct {
	orderStorage        OrderStorage
	productStorage      ProductStorage
	notificationService *NotificationService
	logger              *zap.Logger
}

func NewOrderService(orderStorage OrderStorage, productStorage ProductStorage, notificationService *NotificationService, logger *zap.Logger) *OrderService {
	return &OrderService{
		orderStorage:        orderStorage,
		productStorage:      productStorage,
		notificationService: notificationService,
		logger:              logger,
	}
}

func (s *OrderService) Create(ctx context.Context, input *dto.OrderCreateDTO) (*models.Order, error) {
	s.logger.Info("Creating new order")

	if err := s.validateOrderItems(ctx, input.Items); err != nil {
		return nil, err
	}

	order, err := models.NewOrder(*input)
	if err != nil {
		s.logger.Error("Failed to create order model", zap.Error(err))
		return nil, err
	}

	if err := s.reserveStock(ctx, order.Items); err != nil {
		s.logger.Error("Failed to reserve stock", zap.Error(err))
		return nil, err
	}

	if err := s.orderStorage.Create(ctx, order); err != nil {
		if rollbackErr := s.rollbackStockReservation(ctx, order.Items); rollbackErr != nil {
			s.logger.Error("Failed to rollback stock reservation", zap.Error(rollbackErr))
		}
		s.logger.Error("Failed to create order in storage", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Order created successfully", zap.String("order_id", order.ID), zap.String("order_number", order.OrderNumber))
	
	// Send notification about new order
	if s.notificationService != nil {
		go func() {
			notifyCtx := context.Background()
			if err := s.notificationService.SendOrderCreatedNotification(notifyCtx, order); err != nil {
				s.logger.Error("Failed to send order notification", 
					zap.String("order_id", order.ID),
					zap.Error(err))
			}
		}()
	}
	
	return order, nil
}

func (s *OrderService) GetByID(ctx context.Context, id string) (*models.Order, error) {
	return s.orderStorage.GetByID(ctx, id)
}

func (s *OrderService) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	return s.orderStorage.GetByOrderNumber(ctx, orderNumber)
}

func (s *OrderService) List(ctx context.Context, filter *types.OrderFilter) ([]*models.Order, int64, error) {
	return s.orderStorage.List(ctx, filter)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error {
	s.logger.Info("Updating order status", zap.String("order_id", id), zap.String("status", string(status)))

	order, err := s.orderStorage.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if status == models.OrderStatusCancelled && order.Status != models.OrderStatusCancelled {
		if err := s.rollbackStockReservation(ctx, order.Items); err != nil {
			s.logger.Error("Failed to rollback stock on order cancellation", zap.Error(err))
		}
	}

	if err := s.orderStorage.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	s.logger.Info("Order status updated successfully", zap.String("order_id", id), zap.String("new_status", string(status)))
	return nil
}

func (s *OrderService) Delete(ctx context.Context, id string) error {
	s.logger.Info("Deleting order", zap.String("order_id", id))

	order, err := s.orderStorage.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if order.Status != models.OrderStatusCancelled {
		if err := s.rollbackStockReservation(ctx, order.Items); err != nil {
			s.logger.Error("Failed to rollback stock on order deletion", zap.Error(err))
		}
	}

	if err := s.orderStorage.Delete(ctx, id); err != nil {
		return err
	}

	s.logger.Info("Order deleted successfully", zap.String("order_id", id))
	return nil
}

func (s *OrderService) GetStatistics(ctx context.Context) (map[string]any, error) {
	return s.orderStorage.GetOrderStatistics(ctx)
}

func (s *OrderService) validateOrderItems(ctx context.Context, items []dto.OrderItemDTO) error {
	for i, item := range items {
		product, err := s.productStorage.GetByID(ctx, item.ProductID)
		if err != nil {
			s.logger.Error("Product not found during order validation", zap.Error(err))
			return apperror.New(apperror.CodeNotFound, fmt.Sprintf("product not found for item %d (product_id: %s)", i+1, item.ProductID))
		}

		if !product.IsActive {
			s.logger.Error("Product is not active during order validation")
			return apperror.New(apperror.CodeInvalidInput, fmt.Sprintf("product is not active for item %d", i+1))
		}

		variant, err := s.productStorage.GetVariantByID(ctx, item.ProductID, item.VariantID)
		if err != nil {
			s.logger.Error("Variant not found during order validation", zap.Error(err))
			return apperror.New(apperror.CodeNotFound, fmt.Sprintf("variant not found for item %d (variant_id: %s)", i+1, item.VariantID))
		}

		if variant.Stock < item.Quantity {
			s.logger.Error("Insufficient stock during order validation")
			return apperror.New(apperror.CodeInvalidInput, fmt.Sprintf("insufficient stock for item %d: requested %d, available %d", i+1, item.Quantity, variant.Stock))
		}

		s.logger.Debug("Order item validated successfully")
	}

	return nil
}

func (s *OrderService) reserveStock(ctx context.Context, items []models.OrderItem) error {
	for _, item := range items {
		if err := s.productStorage.UpdateVariantStock(ctx, item.VariantID, -item.Quantity); err != nil {
			return err
		}
	}
	return nil
}

func (s *OrderService) rollbackStockReservation(ctx context.Context, items []models.OrderItem) error {
	for _, item := range items {
		if err := s.productStorage.UpdateVariantStock(ctx, item.VariantID, item.Quantity); err != nil {
			s.logger.Error("Failed to rollback stock for item", 
				zap.String("variant_id", item.VariantID), 
				zap.Int("quantity", item.Quantity), 
				zap.Error(err))
		}
	}
	return nil
}