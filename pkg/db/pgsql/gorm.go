package pgsql

import (
	"caviar/internal/config"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	defaultMaxIdleConns    = 10
	defaultMaxOpenConns    = 100
	defaultConnMaxLifetime = time.Hour
)

type DB struct {
	*gorm.DB
	sqlDB *sql.DB
}

func New(ctx context.Context, cfg config.Postgres, models ...interface{}) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	logLevel := logger.Silent
	if cfg.LogSQL {
		logLevel = logger.Info
	}

	newLogger := logger.New(
		log.New(os.Stdout, "[GORM] ", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, // highlight queries slower than this
			LogLevel:      logLevel,
			Colorful:      true,
		},
	)

	gormConfig := &gorm.Config{
		Logger:         newLogger,
		NamingStrategy: schema.NamingStrategy{SingularTable: false},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get DB from GORM")
	}

	sqlDB.SetMaxIdleConns(defaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(defaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(defaultConnMaxLifetime)

	db = db.WithContext(ctx)

	wrapper := &DB{DB: db, sqlDB: sqlDB}

	if cfg.Migrate && len(models) > 0 {
		if err := wrapper.Migrate(models...); err != nil {
			return nil, errors.Wrap(err, "auto migration failed")
		}
	}

	return wrapper, nil
}

func (d *DB) Close() error {
	return d.sqlDB.Close()
}

func (d *DB) Migrate(models ...interface{}) error {
	return d.DB.AutoMigrate(models...)
}