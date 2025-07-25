package config

import (
	"os"
)

type Config struct {
	MongoDB MongoConfig
	NATS    NATSConfig
	Server  ServerConfig
	Service ServiceConfig
}

type MongoConfig struct {
	URI      string
	Database string
}

type NATSConfig struct {
	URL string
}

type ServerConfig struct {
	Port string
	Host string
}

type ServiceConfig struct {
	Name    string
	Version string
}

func Load() *Config {
	return &Config{
		MongoDB: MongoConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://admin:password@localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "ecommerce"),
		},
		NATS: NATSConfig{
			URL: getEnv("NATS_URL", "nats://localhost:4222"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Service: ServiceConfig{
			Name:    getEnv("SERVICE_NAME", "ecommerce-service"),
			Version: getEnv("SERVICE_VERSION", "1.0.0"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
