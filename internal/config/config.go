package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env string `envconfig:"ENV"`
	Storage
	Cache
	Broker
	Token
	HTTPServer
}

type HTTPServer struct {
	Address         string        `envconfig:"SERVER_ADDRESS"`
	Timeout         time.Duration `envconfig:"SERVER_TIMEOUT"`
	IdleTimeout     time.Duration `envconfig:"SERVER_IDLE_TIMEOUT"`
	ShutdownTimeout time.Duration `envconfig:"SERVER_SHUTDOWN_TIMEOUT"`
}

type Storage struct {
	User     string `envconfig:"STORAGE_USER"`
	Password string `envconfig:"STORAGE_PASSWORD"`
	Host     string `envconfig:"STORAGE_HOST"`
	Port     string `envconfig:"STORAGE_PORT"`
	Database string `envconfig:"STORAGE_DB"`
	SSLMode  string `envconfig:"STORAGE_SSLMODE"`
}

type Cache struct {
	Host string `envconfig:"CACHE_HOST"`
	Port string `envconfig:"CACHE_PORT"`
	DB   int    `envconfig:"CACHE_DB"`
}

type Broker struct {
	URL       string `envconfig:"BROKER_URL"`
	QueueName string `envconfig:"QUEUE_NAME"`
}

type Token struct {
	JWT struct {
		Secret string        `envconfig:"JWT_TOKEN_SECRET"`
		TTL    time.Duration `envconfig:"JWT_TOKEN_TTL"`
	}
	Refresh struct {
		TTL time.Duration `envconfig:"REFRESH_TOKEN_TTL"`
	}
}

func MustLoad() *Config {
	var cfg Config

	err := godotenv.Load()
	if err != nil {
		log.Panicf("failed to load .env file: %v", err)
	}

	err = envconfig.Process("", &cfg)
	if err != nil {
		log.Panicf("failed to make config: %v", err)
	}

	return &cfg
}
