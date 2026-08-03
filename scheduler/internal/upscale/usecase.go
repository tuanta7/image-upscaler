package upscale

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TasksQueue   = "upscale.tasks"
	ResultsQueue = "upscale.results"
)

type UseCase struct {
	jobRepo   *JobRepository
	storage   *Storage
	publisher *amqp.Channel
}

func NewUseCase(
	jobRepo *JobRepository,
	storage *Storage,
	publisher *amqp.Channel,
) *UseCase {
	return &UseCase{
		jobRepo:   jobRepo,
		storage:   storage,
		publisher: publisher,
	}
}

type upscaleTask struct {
	TaskID    string `json:"task_id"`
	InputKey  string `json:"input_key"`
	OutputKey string `json:"output_key"`
	Scale     int    `json:"scale"`
}

func (uc *UseCase) ScheduleTask(ctx context.Context, file []byte, scale int) (string, error) {
	taskID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate task id: %w", err)
	}

	id := taskID.String()
	inputKey := fmt.Sprintf("uploads/%s.png", id)
	outputKey := fmt.Sprintf("results/%s.png", id)

	if err := uc.storage.Upload(ctx, inputKey, file); err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}

	if err := uc.jobRepo.Create(ctx, Job{ID: id, Status: "pending"}); err != nil {
		return "", fmt.Errorf("save job: %w", err)
	}

	body, err := json.Marshal(upscaleTask{
		TaskID:    id,
		InputKey:  inputKey,
		OutputKey: outputKey,
		Scale:     scale,
	})
	if err != nil {
		return "", fmt.Errorf("encode task: %w", err)
	}

	if err := uc.publisher.PublishWithContext(ctx, "", TasksQueue, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return "", fmt.Errorf("publish task: %w", err)
	}

	return id, nil
}

func (uc *UseCase) GetJobStatus(ctx context.Context, taskID string) (string, error) {
	return uc.jobRepo.GetStatus(ctx, taskID)
}

func (uc *UseCase) UpdateJobStatus(ctx context.Context, taskID, status string) error {
	return uc.jobRepo.UpdateStatus(ctx, taskID, status)
}

func (uc *UseCase) GetResultURL(ctx context.Context, taskID string) (string, error) {
	return uc.storage.GetPresignURL(ctx, fmt.Sprintf("results/%s.png", taskID))
}
