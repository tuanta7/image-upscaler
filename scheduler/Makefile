DATABASE_URL ?= postgres://postgres:password@localhost:5432/upscaler?sslmode=disable

start:
	go run ./cmd

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down