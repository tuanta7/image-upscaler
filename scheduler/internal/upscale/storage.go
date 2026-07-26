package upscale

import (
	"bytes"
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const resultURLExpiry = 15 * time.Minute

type Storage struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

func NewStorage(client *s3.Client, bucket string) *Storage {
	return &Storage{client: client, presign: s3.NewPresignClient(client), bucket: bucket}
}

func (s *Storage) Upload(ctx context.Context, key string, data []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	return err
}

// PresignGet returns a temporary URL the browser can fetch the object from
// directly, without proxying the bytes through this service.
func (s *Storage) PresignGet(ctx context.Context, key string) (string, error) {
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(resultURLExpiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
