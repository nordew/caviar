package app

import (
	"caviar/internal/config"
	"caviar/internal/controller/rest"
	"caviar/internal/service"
	"caviar/internal/storage"
	"caviar/pkg/db/pgsql"
	"caviar/pkg/telegram"
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

func MustRun(ctx context.Context) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	gormClient, err := pgsql.New(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	minioClient, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKeyID, cfg.Minio.SecretAccessKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to connect to minio: %v", err)
	}

	userStorage := storage.NewUserStorage(gormClient)

	productStorage := storage.NewProductStorage(gormClient)
	productService := service.NewProductService(productStorage, minioClient, logger)

	telegramService, err := telegram.NewService(cfg.Telegram.Token, userStorage, logger)
	if err != nil {
		log.Fatalf("Failed to create telegram service: %v", err)
	}

	// Create notification service
	notificationService := service.NewNotificationService(userStorage, telegramService, logger)

	orderStorage := storage.NewOrderStorage(gormClient)
	orderService := service.NewOrderService(orderStorage, productStorage, notificationService, logger)

	handler := rest.NewHandler(
		cfg.Server.Port, 
		productService,
		orderService,
		logger,
		cfg.IsProd,
	)

	go telegramService.Start(ctx)
	
	handler.RegisterAndRun(gin.Default())
}