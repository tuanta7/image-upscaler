DATABASE_URL ?= postgres://postgres:password@localhost:5432/upscaler?sslmode=disable
DOCKER_MIGRATE = docker run --rm --network host -v $(CURDIR)/migrations:/migrations migrate/migrate:v4.18.1

start:
	go run ./cmd

migrate-up:
	$(DOCKER_MIGRATE) -path=/migrations -database "$(DATABASE_URL)" up

migrate-down:
	$(DOCKER_MIGRATE) -path=/migrations -database "$(DATABASE_URL)" down 1