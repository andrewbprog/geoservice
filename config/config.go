package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"time"
)

type DatabaseConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
}

type RedisConfig struct {
	Host string `env:"REDIS_HOST"`
	Port string `env:"REDIS_PORT" envDefault:"6379"`
}

type Config struct {

	// Application
	AppPort string `env:"APP_PORT" envDefault:"8080"`
	AppEnv  string `env:"APP_ENV"`

	// Database
	Database DatabaseConfig

	// Redis
	Redis RedisConfig

	// JWT
	JWTSecret string `env:"JWT_SECRET"`

	// RPC
	RpcPort string `env:"RPC_PORT" envDefault:"9000"`

	// DaData
	DadataAPIKey    string `env:"DADATA_API_KEY"`
	DadataSecretKey string `env:"DADATA_SECRET_KEY"`

	// Cache TTLs
	UserCacheTTL   time.Duration `env:"USER_CACHE_TTL" envDefault:"60m"`
	DadataCacheTTL time.Duration `env:"DADATA_CACHE_TTL" envDefault:"60m"`
}

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}

// GetDatabaseURL returns the database connection string
func (c *Config) GetDatabaseURL() string {
	return "host=" + c.Database.Host +
		" port=" + c.Database.Port +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.Name +
		" sslmode=disable"
}
