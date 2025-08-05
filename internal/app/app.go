package app

import (
	"caviar/internal/config"
	"caviar/internal/controller/rest"
	"caviar/internal/service"
	"caviar/internal/storage"
	"caviar/pkg/db/pgsql"
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

	productStorage := storage.NewProductStorage(gormClient)
	productService := service.NewProductService(productStorage, minioClient, logger)

	handler := rest.NewHandler(
		cfg.Server.Port, 
		cfg.Auth.Secret, 
		productService,
		cfg.IsProd,
	)
	
	handler.RegisterAndRun(gin.Default())
}