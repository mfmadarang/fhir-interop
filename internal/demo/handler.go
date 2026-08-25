package demo

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/mfmadarang/fhir-interop/internal/validate"
	"gorm.io/gorm"
)

type Handler struct {
	db     *gorm.DB
	client validate.CodeValidator
	jobs   *JobStore
}

func NewHandler(db *gorm.DB, client validate.CodeValidator) *Handler {
	return &Handler{
		db:     db,
		client: client,
		jobs:   NewJobStore(),
	}
}

func (h *Handler) HandleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "reading request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(data) == 0 {
		http.Error(w, "request body is empty", http.StatusBadRequest)
		return
	}

	job := h.jobs.Create()
	go RunPipeline(job, h.db, h.client, data)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": job.ID})
}

func (h *Handler) HandleStream(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	job, ok := h.jobs.Get(id)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case ev, ok := <-job.Events:
			if !ok {
				h.jobs.Delete(id)
				return
			}
			eventName := "stage"
			if _, isTerm := ev.(TerminologyEvent); isTerm {
				eventName = "terminology"
			}
			payload, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			w.Write([]byte("event: " + eventName + "\ndata: " + string(payload) + "\n\n"))
			flusher.Flush()
		case <-job.Done:
			for len(job.Events) > 0 {
				ev := <-job.Events
				eventName := "stage"
				if _, isTerm := ev.(TerminologyEvent); isTerm {
					eventName = "terminology"
				}
				payload, err := json.Marshal(ev)
				if err != nil {
					continue
				}
				w.Write([]byte("event: " + eventName + "\ndata: " + string(payload) + "\n\n"))
				flusher.Flush()
			}
			w.Write([]byte("event: complete\ndata: {}\n\n"))
			flusher.Flush()
			h.jobs.Delete(id)
			return
		case <-r.Context().Done():
			return
		}
	}
}
