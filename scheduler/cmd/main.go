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

	"github.com/tuanta7/image-upscaler/scheduler/internal/config"
	"github.com/tuanta7/image-upscaler/scheduler/internal/transport"
	"github.com/tuanta7/image-upscaler/scheduler/internal/upscale"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	db, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close(ctx)

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

	// Declare here so the queue exists even if a worker hasn't started
	// consuming yet. Otherwise, the default exchange drops the publication.
	_, err = channel.QueueDeclare(upscale.Queue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("declare queue: %v", err)
	}

	if _, err := channel.QueueDeclare(upscale.ResultsQueue, true, false, false, false, nil); err != nil {
		log.Fatalf("declare results queue: %v", err)
	}

	repo := upscale.NewRepository(db)
	storage := upscale.NewStorage(s3Client, cfg.S3Bucket)
	uc := upscale.NewUseCase(repo, storage, channel)

	results, err := initConsumer(conn)
	if err != nil {
		log.Fatalf("consume results queue: %v", err)
	}
	go uc.ConsumeResults(ctx, results)

	handler := upscale.NewHandler(uc)
	router := transport.NewRouter(handler)

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
