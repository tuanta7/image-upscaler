package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tuanta7/image-upscaler/scheduler/internal/transport/rest"

	"github.com/tuanta7/image-upscaler/scheduler/internal/config"
	"github.com/tuanta7/image-upscaler/scheduler/internal/transport"
	"github.com/tuanta7/image-upscaler/scheduler/internal/upscale"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	s3Client := s3.New(s3.Options{
		BaseEndpoint: aws.String(cfg.S3Endpoint),
		Region:       cfg.S3Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKey,
			cfg.S3SecretKey,
			"",
		),
		UsePathStyle: true,
	})

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("connect rabbitmq: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("open channel: %v", err)
	}
	defer channel.Close()

	if err = declareQueue(channel, upscale.TasksQueue); err != nil {
		log.Fatalf("declare queue: %v", err)
	}

	if err = declareQueue(channel, upscale.ResultsQueue); err != nil {
		log.Fatalf("declare results queue: %v", err)
	}

	dbConn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer dbConn.Close(ctx)

	repo := upscale.NewJobRepository(dbConn)
	storage := upscale.NewStorage(s3Client, cfg.S3Bucket)
	uc := upscale.NewUseCase(repo, storage, channel)
	handler := rest.NewUpscaleHandler(uc)
	router := transport.NewRouter(handler)

	go func() {
		err = consumeResults(conn, nil)
		if err != nil {
			log.Fatalf("consume results queue: %v", err)
		}
	}()

	log.Printf("listening on %s", cfg.BindAddr)
	if err := http.ListenAndServe(cfg.BindAddr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
