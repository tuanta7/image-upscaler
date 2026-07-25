package image

import "context"

type Storage struct{}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Save(ctx context.Context, name string, data []byte) error {
	return nil
}

func (s *Storage) Get(ctx context.Context, name string) ([]byte, error) {
	return nil, nil
}
