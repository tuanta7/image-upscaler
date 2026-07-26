package upscale

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// terminal statuses stop an SSE stream and unlock the result fetch.
func isTerminal(status string) bool {
	return status == "done" || status == "failed"
}

type Handler struct {
	uc *UseCase
}

func NewHandler(uc *UseCase) *Handler {
	return &Handler{uc: uc}
}

type Request struct {
	Image []byte
	Scale int
}

func (h *Handler) Upscale(w http.ResponseWriter, r *http.Request) {
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

	taskID, err := h.uc.Schedule(r.Context(), Request{Image: image, Scale: scale})
	if err != nil {
		http.Error(w, "failed to schedule task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"task_id": taskID})
}

// Events streams a task's status over SSE until it reaches a terminal
// status or the client disconnects.
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	taskID := chi.URLParam(r, "id")
	ch, unsubscribe := h.uc.Subscribe(taskID)
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Send the current status right away in case it already changed (or
	// finished) before this subscriber connected.
	status, err := h.uc.Status(r.Context(), taskID)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", status)
	flusher.Flush()
	if isTerminal(status) {
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case status := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", status)
			flusher.Flush()
			if isTerminal(status) {
				return
			}
		}
	}
}

// Result returns a presigned URL for the upscaled image of a finished task.
func (h *Handler) Result(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	url, err := h.uc.ResultURL(r.Context(), taskID)
	if err != nil {
		http.Error(w, "result not ready", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}
