package storage

import (
	"context"

	"caviar/internal/models"
	"caviar/internal/types"
	"caviar/pkg/apperror"

	"gorm.io/gorm"
)

type OrderStorage struct {
	db *gorm.DB
}

func NewOrderStorage(db *gorm.DB) *OrderStorage {
	return &OrderStorage{
		db: db,
	}
}

func (s *OrderStorage) Create(ctx context.Context, order *models.Order) error {
	if err := s.db.WithContext(ctx).Create(order).Error; err != nil {
		return apperror.New(apperror.CodeInternal, "failed to create order: "+err.Error())
	}
	return nil
}

func (s *OrderStorage) GetByID(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order
	err := s.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Where("id = ?", id).
		First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.CodeNotFound, "order not found")
		}
		return nil, apperror.New(apperror.CodeInternal, "failed to get order: "+err.Error())
	}

	return &order, nil
}

func (s *OrderStorage) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	var order models.Order
	err := s.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Where("order_number = ?", orderNumber).
		First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.CodeNotFound, "order not found")
		}
		return nil, apperror.New(apperror.CodeInternal, "failed to get order: "+err.Error())
	}

	return &order, nil
}

func (s *OrderStorage) List(ctx context.Context, filter *types.OrderFilter) ([]*models.Order, int64, error) {
	var orders []*models.Order
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Order{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.CustomerPhone != "" {
		query = query.Where("customer_info->>'phone' ILIKE ?", "%"+filter.CustomerPhone+"%")
	}

	if filter.Country != "" {
		query = query.Where("delivery_info->>'country' = ?", filter.Country)
	}

	if !filter.CreatedFrom.IsZero() {
		query = query.Where("created_at >= ?", filter.CreatedFrom)
	}

	if !filter.CreatedTo.IsZero() {
		query = query.Where("created_at <= ?", filter.CreatedTo)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperror.New(apperror.CodeInternal, "failed to count orders: "+err.Error())
	}

	query = query.Preload("Items").
		Preload("Items.Product").
		Order("created_at DESC")

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&orders).Error; err != nil {
		return nil, 0, apperror.New(apperror.CodeInternal, "failed to list orders: "+err.Error())
	}

	return orders, total, nil
}

func (s *OrderStorage) Update(ctx context.Context, order *models.Order) error {
	if err := s.db.WithContext(ctx).Save(order).Error; err != nil {
		return apperror.New(apperror.CodeInternal, "failed to update order: "+err.Error())
	}
	return nil
}

func (s *OrderStorage) UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error {
	result := s.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"updated_at": "NOW()",
		})

	if result.Error != nil {
		return apperror.New(apperror.CodeInternal, "failed to update order status: "+result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return apperror.New(apperror.CodeNotFound, "order not found")
	}

	return nil
}

func (s *OrderStorage) Delete(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&models.Order{}, "id = ?", id)

	if result.Error != nil {
		return apperror.New(apperror.CodeInternal, "failed to delete order: "+result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return apperror.New(apperror.CodeNotFound, "order not found")
	}

	return nil
}

func (s *OrderStorage) GetOrderStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalOrders int64
	if err := s.db.WithContext(ctx).Model(&models.Order{}).Count(&totalOrders).Error; err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to count total orders: "+err.Error())
	}
	stats["total_orders"] = totalOrders

	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := s.db.WithContext(ctx).
		Model(&models.Order{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusCounts).Error; err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to get status counts: "+err.Error())
	}

	statusMap := make(map[string]int64)
	for _, sc := range statusCounts {
		statusMap[sc.Status] = sc.Count
	}
	stats["status_counts"] = statusMap

	var countryCounts []struct {
		Country string
		Count   int64
	}
	if err := s.db.WithContext(ctx).
		Model(&models.Order{}).
		Select("delivery_info->>'country' as country, count(*) as count").
		Group("delivery_info->>'country'").
		Find(&countryCounts).Error; err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to get country counts: "+err.Error())
	}

	countryMap := make(map[string]int64)
	for _, cc := range countryCounts {
		countryMap[cc.Country] = cc.Count
	}
	stats["country_counts"] = countryMap

	return stats, nil
}