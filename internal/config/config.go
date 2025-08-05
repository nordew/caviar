package config

import (
	"time"
)

type Config struct {
	Postgres        Postgres        `envPrefix:"POSTGRES_"`
	Minio           Minio           `envPrefix:"MINIO_"`
	Server          Server          `envPrefix:"SERVER_"`
	JWT             JWT             `envPrefix:"JWT_"`
	Logger          Logger          `envPrefix:"LOGGER_"`
	HTTP            HTTP            `envPrefix:"HTTP_"`
	RateLimiter     RateLimiter     `envPrefix:"RATE_LIMITER_"`
	Auth Auth `envPrefix:"AUTH_"`
	IsProd bool `env:"IS_PROD" envDefault:"false"`
}

type Auth struct {
	Secret string `env:"SECRET,required"`
}

type Postgres struct {
	Host     string `env:"HOST,required"`
	Port     string `env:"PORT,required"`
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	DBName   string `env:"DBNAME,required"`
	Migrate  bool   `env:"MIGRATE" envDefault:"false"`
	LogSQL   bool   `env:"LOG_SQL" envDefault:"false"`
}

type Server struct {
	Port    string        `env:"PORT" envDefault:"8080"`
	Timeout time.Duration `env:"TIMEOUT" envDefault:"10s"`
}

type JWT struct {
	SecretKey       string        `env:"SECRET_KEY,required"`
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"15m"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" envDefault:"24h"`
}

type Logger struct {
	Level string `env:"LEVEL" envDefault:"debug"`
}

type HTTP struct {
	Host               string   `env:"HOST" envDefault:"localhost"`
	Port               string   `env:"PORT" envDefault:"8080"`
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:","`
}

type RateLimiter struct {
	RPS   int           `env:"RPS" envDefault:"10"`
	Burst int           `env:"BURST" envDefault:"20"`
	TTL   time.Duration `env:"TTL" envDefault:"10m"`
}

type Minio struct {
	Endpoint        string `env:"ENDPOINT,required"`
	AccessKeyID     string `env:"ACCESS_KEY_ID,required"`
	SecretAccessKey string `env:"SECRET_ACCESS_KEY,required"`
	BucketName      string `env:"BUCKET_NAME,required"`
	UseSSL          bool   `env:"USE_SSL" envDefault:"false"`
}
