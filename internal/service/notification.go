package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"caviar/internal/models"

	"go.uber.org/zap"
)

const (
	batchSize                = 30
	batchDelay               = 100 * time.Millisecond
	dateTimeFormat          = "02.01.2006 15:04"
	
	msgNewOrder             = "🆕 <b>Нове замовлення!</b>\n\n"
	msgOrderNumber          = "📋 <b>Номер:</b> %s\n"
	msgOrderAmount          = "💰 <b>Сума:</b> %d %s\n"
	msgOrderDate            = "📅 <b>Дата:</b> %s\n"
	msgCustomer             = "👤 <b>Клієнт:</b> %s\n"
	msgPhone                = "📱 <b>Телефон:</b> %s\n"
	msgDelivery             = "🚚 <b>Доставка:</b> %s\n"
	msgLocation             = "🌍 <b>Місто:</b> %s, %s\n"
	msgPostOffice           = "📦 <b>Відділення:</b> %s\n"
	msgAddress              = "🏠 <b>Адреса:</b> %s\n"
	msgItemsCount           = "\n📦 <b>Товарів:</b> %d шт.\n"
	msgNotes                = "\n💭 <b>Примітки:</b> %s\n"
	
	deliveryPostOffice      = "Нова пошта"
	deliveryCourier         = "Кур'єрська доставка"
	deliveryAddress         = "За адресою"
)

type NotificationService struct {
	userStorage      UserStorage
	telegramService  TelegramNotifier
	logger          *zap.Logger
	enabledChannels []NotificationChannel
}

type NotificationChannel string

const (
	ChannelTelegram NotificationChannel = "telegram"
	ChannelEmail    NotificationChannel = "email"
	ChannelSMS      NotificationChannel = "sms"
)

type UserStorage interface {
	GetWithTelegramID(ctx context.Context) ([]*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
}

type TelegramNotifier interface {
	SendMessage(ctx context.Context, telegramID int64, message string) error
	SendMessageToMultiple(ctx context.Context, telegramIDs []int64, message string) error
}

type NotificationRequest struct {
	Title     string
	Message   string
	Channels  []NotificationChannel
	UserIDs   []string
	Priority  Priority
	Metadata  map[string]any
}

type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

type NotificationResult struct {
	Channel     NotificationChannel
	Success     int
	Failed      int
	Errors      []error
	ProcessedAt time.Time
}

func NewNotificationService(
	userStorage UserStorage,
	telegramService TelegramNotifier,
	logger *zap.Logger,
) *NotificationService {
	return &NotificationService{
		userStorage:     userStorage,
		telegramService: telegramService,
		logger:          logger,
		enabledChannels: []NotificationChannel{ChannelTelegram},
	}
}

func (s *NotificationService) SetEnabledChannels(channels []NotificationChannel) {
	s.enabledChannels = channels
	s.logger.Info("Updated enabled notification channels",
		zap.Strings("channels", channelsToStrings(channels)))
}

func (s *NotificationService) SendOrderCreatedNotification(ctx context.Context, order *models.Order) error {
	message := s.formatOrderCreatedMessage(order)
	
	req := &NotificationRequest{
		Title:    "Нове замовлення",
		Message:  message,
		Channels: []NotificationChannel{ChannelTelegram},
		Priority: PriorityNormal,
		Metadata: map[string]any{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
			"total_amount": order.TotalAmount,
		},
	}
	
	return s.SendNotification(ctx, req)
}

func (s *NotificationService) SendNotification(ctx context.Context, req *NotificationRequest) error {
	s.logger.Info("Sending notification",
		zap.String("title", req.Title),
		zap.Strings("channels", channelsToStrings(req.Channels)),
		zap.Int("priority", int(req.Priority)))
	
	var results []NotificationResult
	var lastError error
	
	for _, channel := range req.Channels {
		if !s.isChannelEnabled(channel) {
			s.logger.Debug("Channel not enabled, skipping",
				zap.String("channel", string(channel)))
			continue
		}
		
		result, err := s.sendToChannel(ctx, channel, req)
		if err != nil {
			s.logger.Error("Failed to send notification to channel",
				zap.String("channel", string(channel)),
				zap.Error(err))
			lastError = err
		}
		
		results = append(results, result)
	}
	
	s.logNotificationResults(req, results)
	return lastError
}

func (s *NotificationService) sendToChannel(ctx context.Context, channel NotificationChannel, req *NotificationRequest) (NotificationResult, error) {
	result := NotificationResult{
		Channel:     channel,
		ProcessedAt: time.Now(),
	}
	
	switch channel {
	case ChannelTelegram:
		return s.sendTelegramNotification(ctx, req)
	case ChannelEmail:
		result.Errors = []error{fmt.Errorf("email notifications not implemented")}
		return result, fmt.Errorf("email notifications not implemented")
	case ChannelSMS:
		result.Errors = []error{fmt.Errorf("SMS notifications not implemented")}
		return result, fmt.Errorf("SMS notifications not implemented")
	default:
		err := fmt.Errorf("unknown notification channel: %s", channel)
		result.Errors = []error{err}
		return result, err
	}
}

func (s *NotificationService) sendTelegramNotification(ctx context.Context, req *NotificationRequest) (NotificationResult, error) {
	result := NotificationResult{
		Channel:     ChannelTelegram,
		ProcessedAt: time.Now(),
	}
	
	users, err := s.getTargetUsers(ctx, req.UserIDs)
	if err != nil {
		result.Errors = []error{err}
		return result, fmt.Errorf("failed to get target users: %w", err)
	}
	
	if len(users) == 0 {
		s.logger.Info("No users with Telegram ID found for notification")
		return result, nil
	}
	
	var telegramIDs []int64
	userMap := make(map[int64]*models.User)
	
	for _, user := range users {
		if user.TelegramID != nil && *user.TelegramID > 0 {
			telegramIDs = append(telegramIDs, *user.TelegramID)
			userMap[*user.TelegramID] = user
		}
	}
	
	if len(telegramIDs) == 0 {
		s.logger.Info("No users with valid Telegram ID found")
		return result, nil
	}
	
	s.logger.Info("Sending Telegram notification to users",
		zap.Int("user_count", len(telegramIDs)))
	
	successCount, failureCount, errors := s.sendTelegramBatch(ctx, telegramIDs, req.Message, userMap)
	
	result.Success = successCount
	result.Failed = failureCount
	result.Errors = errors
	
	return result, nil
}

func (s *NotificationService) sendTelegramBatch(ctx context.Context, telegramIDs []int64, message string, userMap map[int64]*models.User) (int, int, []error) {
	var successCount, failureCount int
	var errors []error
	
	for i := 0; i < len(telegramIDs); i += batchSize {
		end := i + batchSize
		if end > len(telegramIDs) {
			end = len(telegramIDs)
		}
		
		batch := telegramIDs[i:end]
		
		for _, telegramID := range batch {
			user := userMap[telegramID]
			
			err := s.telegramService.SendMessage(ctx, telegramID, message)
			if err != nil {
				s.logger.Warn("Failed to send Telegram message to user",
					zap.Int64("telegram_id", telegramID),
					zap.String("user_id", user.ID.String()),
					zap.Error(err))
				failureCount++
				errors = append(errors, fmt.Errorf("user %s: %w", user.ID.String(), err))
			} else {
				successCount++
				s.logger.Debug("Telegram message sent successfully",
					zap.Int64("telegram_id", telegramID),
					zap.String("user_id", user.ID.String()))
			}
		}
		
		if i+batchSize < len(telegramIDs) {
			time.Sleep(batchDelay)
		}
	}
	
	return successCount, failureCount, errors
}

func (s *NotificationService) getTargetUsers(ctx context.Context, userIDs []string) ([]*models.User, error) {
	if len(userIDs) == 0 {
		return s.userStorage.GetWithTelegramID(ctx)
	}
	
	var users []*models.User
	for _, userID := range userIDs {
		user, err := s.userStorage.GetByID(ctx, userID)
		if err != nil {
			s.logger.Warn("Failed to get user by ID",
				zap.String("user_id", userID),
				zap.Error(err))
			continue
		}
		users = append(users, user)
	}
	
	return users, nil
}

func (s *NotificationService) formatOrderCreatedMessage(order *models.Order) string {
	var sb strings.Builder
	
	sb.WriteString(msgNewOrder)
	sb.WriteString(fmt.Sprintf(msgOrderNumber, order.OrderNumber))
	sb.WriteString(fmt.Sprintf(msgOrderAmount, order.TotalAmount.Amount, order.TotalAmount.Currency))
	sb.WriteString(fmt.Sprintf(msgOrderDate, order.CreatedAt.Format(dateTimeFormat)))
	
	if order.CustomerInfo.FullName != "" {
		sb.WriteString(fmt.Sprintf(msgCustomer, order.CustomerInfo.FullName))
	} else if order.CustomerInfo.FirstName != "" {
		name := order.CustomerInfo.FirstName
		if order.CustomerInfo.LastName != "" {
			name += " " + order.CustomerInfo.LastName
		}
		sb.WriteString(fmt.Sprintf(msgCustomer, name))
	}
	
	if order.CustomerInfo.Phone != "" {
		sb.WriteString(fmt.Sprintf(msgPhone, order.CustomerInfo.Phone))
	}
	
	sb.WriteString(fmt.Sprintf(msgDelivery, s.getDeliveryTypeUkrainian(order.DeliveryInfo.Type)))
	sb.WriteString(fmt.Sprintf(msgLocation, order.DeliveryInfo.City, order.DeliveryInfo.Country))
	
	if order.DeliveryInfo.PostOffice != "" {
		sb.WriteString(fmt.Sprintf(msgPostOffice, order.DeliveryInfo.PostOffice))
	}
	if order.DeliveryInfo.Address != "" {
		sb.WriteString(fmt.Sprintf(msgAddress, order.DeliveryInfo.Address))
	}
	
	sb.WriteString(fmt.Sprintf(msgItemsCount, len(order.Items)))
	
	if order.Notes != "" {
		sb.WriteString(fmt.Sprintf(msgNotes, order.Notes))
	}
	
	return sb.String()
}

func (s *NotificationService) getDeliveryTypeUkrainian(deliveryType models.DeliveryType) string {
	switch deliveryType {
	case models.DeliveryTypePostOffice:
		return deliveryPostOffice
	case models.DeliveryTypeCourier:
		return deliveryCourier
	case models.DeliveryTypeAddress:
		return deliveryAddress
	default:
		return string(deliveryType)
	}
}

func (s *NotificationService) isChannelEnabled(channel NotificationChannel) bool {
	for _, enabled := range s.enabledChannels {
		if enabled == channel {
			return true
		}
	}
	return false
}

func (s *NotificationService) logNotificationResults(req *NotificationRequest, results []NotificationResult) {
	for _, result := range results {
		if len(result.Errors) > 0 {
			s.logger.Error("Notification channel had errors",
				zap.String("channel", string(result.Channel)),
				zap.Int("success", result.Success),
				zap.Int("failed", result.Failed),
				zap.Int("error_count", len(result.Errors)))
		} else {
			s.logger.Info("Notification sent successfully",
				zap.String("channel", string(result.Channel)),
				zap.Int("success", result.Success),
				zap.Int("failed", result.Failed))
		}
	}
}

func channelsToStrings(channels []NotificationChannel) []string {
	var result []string
	for _, channel := range channels {
		result = append(result, string(channel))
	}
	return result
}