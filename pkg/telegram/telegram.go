package telegram

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"caviar/internal/models"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type otpStore struct {
    codes map[string]string
    times map[string]time.Time
    mu    sync.RWMutex
}

type UserStorage interface {
    GetByTelegramID(ctx context.Context, telegramID string) (*models.User, error)
}

type Service struct {
    bot         *tele.Bot
    store       *otpStore
    userStorage UserStorage
    logger      *zap.Logger
}

func newOTPStore() *otpStore {
    store := &otpStore{
        codes: make(map[string]string),
        times: make(map[string]time.Time),
    }

    go func() {
        ticker := time.NewTicker(10 * time.Minute)
        defer ticker.Stop()

        for range ticker.C {
            store.cleanup()
        }
    }()

    return store
}

func NewService(
	token string,
	storage UserStorage,
	logger *zap.Logger,
) (*Service, error) {
    bot, err := tele.NewBot(tele.Settings{
        Token:  token,
        Poller: &tele.LongPoller{Timeout: 10 * time.Second},
    })
    if err != nil {
        return nil, err
    }

    s := &Service{
        bot:         bot,
        store:       newOTPStore(),
        userStorage: storage,
        logger:      logger,
    }

    s.setupHandlers()

    return s, nil
}

func (s *otpStore) storeCode(id string, code string) {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.codes[id] = code
    s.times[id] = time.Now()
}

func (s *otpStore) verifyCode(id string, code string) bool {
    now := time.Now()

    s.mu.Lock()
    defer s.mu.Unlock()

    stored, ok := s.codes[id]
    ts, okTime := s.times[id]
    if !ok || !okTime || now.Sub(ts) > time.Hour || stored != code {
        delete(s.codes, id)
        delete(s.times, id)
        return false
    }

    delete(s.codes, id)
    delete(s.times, id)
    return true
}

func (s *otpStore) cleanup() {
    now := time.Now()

    s.mu.Lock()
    for id, ts := range s.times {
        if now.Sub(ts) > time.Hour {
            delete(s.codes, id)
            delete(s.times, id)
        }
    }
    s.mu.Unlock()
}



func (s *Service) setupHandlers() {
    s.bot.Handle("/start", s.handleStart)
    s.bot.Handle("/login", s.handleLogin)
}

func (s *Service) Start(ctx context.Context) {
    go s.bot.Start()
    <-ctx.Done()
    s.bot.Stop()
}

func (s *Service) handleStart(c tele.Context) error {
    return c.Send("👋 Welcome! Use /login to get your login code.")
}

func (s *Service) handleLogin(c tele.Context) error {
    id := c.Sender().ID

    if _, err := s.userStorage.GetByTelegramID(context.Background(), strconv.FormatInt(id, 10)); err != nil {
        return c.Send("❌ You're not registered. Contact admin to link your account.")
    }

    code := s.generateCode(strconv.FormatInt(id, 10))
    message := fmt.Sprintf("🔐 Your login code: <code>%s</code>\n⏰ Expires in 1 hour", code)

    return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeHTML})
}

func (s *Service) RequestOTP(ctx context.Context, telegramID string) error {
    if _, err := s.userStorage.GetByTelegramID(ctx, telegramID); err != nil {
        return ErrUserNotFound
    }

    code := s.generateCode(telegramID)
    message := fmt.Sprintf("🔐 Your login code: <code>%s</code>\n⏰ Expires in 1 hour", code)

    telegramIDInt, _ := strconv.ParseInt(telegramID, 10, 64)
    _, err := s.bot.Send(&tele.User{ID: telegramIDInt}, message, &tele.SendOptions{ParseMode: tele.ModeHTML})
    return err
}

func (s *Service) ValidateOTP(ctx context.Context, telegramID string, code string) (*models.User, error) {
    if !s.store.verifyCode(telegramID, code) {
        return nil, ErrOTPInvalid
    }

    usr, err := s.userStorage.GetByTelegramID(ctx, telegramID)
    if err != nil {
        return nil, ErrUserNotFound
    }

    return &models.User{
        ID:        usr.ID,
        Email:     usr.Email,
        FirstName: usr.FirstName,
        LastName:  usr.LastName,
        IsActive:  usr.IsActive,
        CreatedAt: usr.CreatedAt,
    }, nil
}

func (s *Service) generateCode(id string) string {
    code := fmt.Sprintf("%06d", rand.Intn(1e6))
    s.store.storeCode(id, code)
    return code
}

func (s *Service) GetUserByTelegramID(ctx context.Context, telegramID string) (*models.User, error) {
    usr, err := s.userStorage.GetByTelegramID(ctx, telegramID)
    if err != nil {
        return nil, err
    }

    return &models.User{
        ID:        usr.ID,
        Email:     usr.Email,
        FirstName: usr.FirstName,
        LastName:  usr.LastName,
        IsActive:  usr.IsActive,
        CreatedAt: usr.CreatedAt,
    }, nil
}

// SendMessage sends a message to a specific Telegram user
func (s *Service) SendMessage(ctx context.Context, telegramID int64, message string) error {
	_, err := s.bot.Send(&tele.User{ID: telegramID}, message, &tele.SendOptions{
		ParseMode: tele.ModeHTML,
	})
	return err
}

// SendMessageToMultiple sends a message to multiple Telegram users
func (s *Service) SendMessageToMultiple(ctx context.Context, telegramIDs []int64, message string) error {
	var errors []error
	
	for _, telegramID := range telegramIDs {
		err := s.SendMessage(ctx, telegramID, message)
		if err != nil {
			s.logger.Warn("Failed to send message to user",
				zap.Int64("telegram_id", telegramID),
				zap.Error(err))
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to send to %d users", len(errors))
	}
	
	return nil
}
