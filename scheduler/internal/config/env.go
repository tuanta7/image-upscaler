package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type EnvConfig struct {
	BindAddr    string `envconfig:"BIND_ADDR" default:":7031"`
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	RabbitMQURL string `envconfig:"RABBITMQ_URL" required:"true"`
	S3Endpoint  string `envconfig:"S3_ENDPOINT" required:"true"`
	S3Region    string `envconfig:"S3_REGION" default:"us-east-1"`
	S3AccessKey string `envconfig:"S3_ACCESS_KEY" required:"true"`
	S3SecretKey string `envconfig:"S3_SECRET_KEY" required:"true"`
	S3Bucket    string `envconfig:"S3_BUCKET" default:"images"`
}

func Load() *EnvConfig {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment")
	}

	var cfg EnvConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("load config: %v", err)
	}
	return &cfg
}
