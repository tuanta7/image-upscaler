package upscale

import (
	"bytes"
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	client  *s3.Client
	bucket  string
	presign *s3.PresignClient
}

func NewStorage(client *s3.Client, bucket string) *Storage {
	return &Storage{
		client:  client,
		bucket:  bucket,
		presign: s3.NewPresignClient(client),
	}
}

func (s *Storage) Upload(ctx context.Context, key string, data []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	return err
}

func (s *Storage) GetPresignURL(ctx context.Context, key string) (string, error) {
	req, err := s.presign.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(15*time.Minute),
	)
	if err != nil {
		return "", err
	}

	return req.URL, nil
}
