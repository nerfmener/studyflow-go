package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nerfmener/studyflow-go/internal/store"
)

type Server struct {
	store *store.Store
}

func NewServer(studyStore *store.Store) *Server {
	return &Server{store: studyStore}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.home)
	mux.HandleFunc("/api/tasks", s.tasks)
	mux.HandleFunc("/api/tasks/", s.task)
	mux.HandleFunc("/api/sessions", s.sessions)
	mux.HandleFunc("/health", s.health)
	return logging(mux)
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.ParseFiles("web/index.html"))
	snapshot, err := s.store.Snapshot()
	if err != nil {
		http.Error(w, "could not load study data", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, snapshot); err != nil {
		http.Error(w, "could not render page", http.StatusInternalServerError)
	}
}

func (s *Server) tasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tasks, err := s.store.Tasks()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, tasks)
	case http.MethodPost:
		var input struct {
			Title    string `json:"title"`
			Category string `json:"category"`
		}
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		task, err := s.store.AddTask(strings.TrimSpace(input.Title), strings.TrimSpace(input.Category))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, task)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) task(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, err := taskID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid task id"))
		return
	}
	task, err := s.store.ToggleTask(id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrTaskNotFound) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) sessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		TaskID  int    `json:"task_id"`
		Minutes int    `json:"minutes"`
		Note    string `json:"note"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	session, err := s.store.AddSession(input.TaskID, input.Minutes, strings.TrimSpace(input.Note))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func decodeJSON(r *http.Request, destination any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func taskID(r *http.Request) (int, error) {
	value := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	return strconv.Atoi(value)
}
