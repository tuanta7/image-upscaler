package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tuanta7/image-upscaler/scheduler/internal/transport/sse"
)

type UpscaleUC interface {
	ScheduleTask(ctx context.Context, file []byte, scale int) (string, error)
	GetJobStatus(ctx context.Context, taskID string) (string, error)
	UpdateJobStatus(ctx context.Context, taskID, status string) error
	GetResultURL(ctx context.Context, taskID string) (string, error)
}

type UpscaleHandler struct {
	upscaler UpscaleUC
	sseHub   *sse.Hub
}

func NewUpscaleHandler(uc UpscaleUC) *UpscaleHandler {
	return &UpscaleHandler{
		upscaler: uc,
		sseHub:   sse.NewHub(),
	}
}

// PublishStatus fans a status update out to any SSE client watching the task.
func (h *UpscaleHandler) PublishStatus(taskID, status string) {
	h.sseHub.Publish(taskID, status)
}

func (h *UpscaleHandler) Upscale(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "invalid form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	image, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read image", http.StatusBadRequest)
		return
	}

	scale, _ := strconv.Atoi(r.FormValue("scale"))

	taskID, err := h.upscaler.ScheduleTask(r.Context(), image, scale)
	if err != nil {
		http.Error(w, "failed to schedule task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"task_id": taskID})
}

// Events streams a task's status over SSE
// until it reaches a terminal status or the client disconnects.
func (h *UpscaleHandler) Events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	taskID := chi.URLParam(r, "id")

	statusCh, unsubscribe := h.sseHub.Subscribe(taskID)
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	status, err := h.upscaler.GetJobStatus(r.Context(), taskID)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	_, err = fmt.Fprintf(w, "data: %s\n\n", status)
	if err != nil {
		log.Printf("failed to write event: %v", err)
	}

	flusher.Flush()
	if status == "done" || status == "failed" {
		// stop and unsubscribe
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case newStatus := <-statusCh:
			_, _ = fmt.Fprintf(w, "data: %s\n\n", newStatus)
			flusher.Flush()
			if newStatus == "done" || newStatus == "failed" {
				return
			}
		}
	}
}

func (h *UpscaleHandler) Result(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	url, err := h.upscaler.GetResultURL(r.Context(), taskID)
	if err != nil {
		http.Error(w, "result not ready", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
}
