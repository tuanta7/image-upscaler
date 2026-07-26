package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tuanta7/image-upscaler/scheduler/internal/upscale"
	"github.com/tuanta7/image-upscaler/scheduler/web"
)

func NewRouter(imageHandler *upscale.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/upscale", imageHandler.Upscale)
	r.Get("/tasks/{id}/events", imageHandler.Events)
	r.Get("/tasks/{id}/result", imageHandler.Result)
	r.Handle("/*", http.FileServer(http.FS(web.FS)))
	return r
}
