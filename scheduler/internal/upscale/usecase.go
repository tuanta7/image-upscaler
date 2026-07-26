package upscale

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Queue        = "upscale.tasks"
	ResultsQueue = "upscale.results"
)

type UseCase struct {
	repo      *Repository
	storage   *Storage
	publisher *amqp.Channel
	hub       *Hub
}

func NewUseCase(repo *Repository, storage *Storage, publisher *amqp.Channel) *UseCase {
	return &UseCase{
		repo:      repo,
		storage:   storage,
		publisher: publisher,
		hub:       NewHub(),
	}
}

// upscaleTask is the message body a worker expects on the queue.
type upscaleTask struct {
	TaskID    string `json:"task_id"`
	InputKey  string `json:"input_key"`
	OutputKey string `json:"output_key"`
	Scale     int    `json:"scale"`
}

// Schedule uploads the image, records the job, and drops a task on the
// queue for a worker to pick up.
// Fire-and-forget: no retry, no dedup.
func (uc *UseCase) Schedule(ctx context.Context, req Request) (string, error) {
	taskID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate task id: %w", err)
	}

	id := taskID.String()
	inputKey := fmt.Sprintf("uploads/%s.png", id)
	outputKey := fmt.Sprintf("results/%s.png", id)

	if err := uc.storage.Upload(ctx, inputKey, req.Image); err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}

	if err := uc.repo.CreateJob(ctx, Job{ID: id, Status: "pending"}); err != nil {
		return "", fmt.Errorf("save job: %w", err)
	}

	body, err := json.Marshal(upscaleTask{
		TaskID:    id,
		InputKey:  inputKey,
		OutputKey: outputKey,
		Scale:     req.Scale,
	})
	if err != nil {
		return "", fmt.Errorf("encode task: %w", err)
	}

	if err := uc.publisher.PublishWithContext(ctx, "", Queue, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		}); err != nil {
		return "", fmt.Errorf("publish task: %w", err)
	}

	return id, nil
}

// Subscribe streams status updates for a task to the caller (an SSE
// handler). Call the returned func once done reading to release it.
func (uc *UseCase) Subscribe(taskID string) (<-chan string, func()) {
	return uc.hub.Subscribe(taskID)
}

// Status returns the task's last known status, for callers that connect
// to Subscribe after the status has already changed.
func (uc *UseCase) Status(ctx context.Context, taskID string) (string, error) {
	return uc.repo.GetStatus(ctx, taskID)
}

// ResultURL returns a temporary URL for the upscaled image of a finished
// task, for the browser to fetch directly from S3.
func (uc *UseCase) ResultURL(ctx context.Context, taskID string) (string, error) {
	return uc.storage.PresignGet(ctx, fmt.Sprintf("results/%s.png", taskID))
}

// resultMsg is the message body a worker publishes to ResultsQueue when a
// task's status changes.
type resultMsg struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

// ConsumeResults persists status updates from workers and fans them out to
// any SSE subscribers. Runs until msgs is closed.
func (uc *UseCase) ConsumeResults(ctx context.Context, msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		var res resultMsg
		if err := json.Unmarshal(msg.Body, &res); err != nil {
			log.Printf("discard malformed result message: %v", err)
			continue
		}

		if err := uc.repo.UpdateStatus(ctx, res.TaskID, res.Status); err != nil {
			log.Printf("update status for task %s: %v", res.TaskID, err)
		}
		uc.hub.Publish(res.TaskID, res.Status)
	}
}
