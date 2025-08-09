package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

// Config holds all application configuration
type Config struct {
	MongoDB        MongoConfig   `envPrefix:"MONGODB_"`
	NATS           NATSConfig    `envPrefix:"NATS_"`
	Server         ServerConfig  `envPrefix:"SERVER_"`
	Service        ServiceConfig `envPrefix:"SERVICE_"`
	JWTSecret      string        `env:"JWT_SECRET" envDefault:"JxmnOYhSfqw-g9IkS489eaqZw9uVCzK5H912T9YezJ5MWCHPj4LHo4xOEQixZap38LcpBMuYNUBbgBAH0rTIZQ"`
	Environment    string        `env:"ENVIRONMENT" envDefault:"development"`
	AllowedOrigins []string      `env:"ALLOWED_ORIGINS" envDefault:"*" envSeparator:","`
}

// MongoConfig holds MongoDB config values
type MongoConfig struct {
	URI      string `env:"URI" envDefault:"mongodb://admin:password@localhost:27017"`
	Database string `env:"DATABASE" envDefault:"ecommerce"`
}

// NATSConfig holds NATS connection settings
type NATSConfig struct {
	URL string `env:"URL" envDefault:"nats://localhost:4222"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port string `env:"PORT" envDefault:"8080"`
	Host string `env:"HOST" envDefault:"localhost"`
}

// ServiceConfig holds service metadata
type ServiceConfig struct {
	Name    string `env:"NAME" envDefault:"ecommerce-service"`
	Version string `env:"VERSION" envDefault:"1.0.0"`
}

// Load reads environment variables into Config
func Load() *Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}
	return &cfg
}
