package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrEmptyTitle   = errors.New("task title cannot be empty")
	ErrInvalidMins  = errors.New("study time must be between 1 and 720 minutes")
	ErrTaskNotFound = errors.New("task not found")
)

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

type Session struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	TaskTitle string    `json:"task_title"`
	Minutes   int       `json:"minutes"`
	Note      string    `json:"note"`
	StudiedAt time.Time `json:"studied_at"`
}

type Snapshot struct {
	Tasks          []Task    `json:"tasks"`
	Sessions       []Session `json:"sessions"`
	TotalMinutes   int       `json:"total_minutes"`
	CompletedTasks int       `json:"completed_tasks"`
	Streak         int       `json:"streak"`
}

type fileData struct {
	Tasks    []Task    `json:"tasks"`
	Sessions []Session `json:"sessions"`
}

type Store struct {
	mu   sync.Mutex
	path string
	data fileData
}

func New(path string) (*Store, error) {
	s := &Store{path: path, data: fileData{Tasks: []Task{}, Sessions: []Session{}}}
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(content, &s.data); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Tasks() ([]Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Task(nil), s.data.Tasks...), nil
}

func (s *Store) AddTask(title, category string) (Task, error) {
	if strings.TrimSpace(title) == "" {
		return Task{}, ErrEmptyTitle
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task := Task{ID: nextID(s.data.Tasks), Title: title, Category: category, CreatedAt: time.Now()}
	s.data.Tasks = append(s.data.Tasks, task)
	return task, s.save()
}

func (s *Store) ToggleTask(id int) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.data.Tasks {
		if s.data.Tasks[index].ID == id {
			s.data.Tasks[index].Completed = !s.data.Tasks[index].Completed
			return s.data.Tasks[index], s.save()
		}
	}
	return Task{}, ErrTaskNotFound
}

func (s *Store) AddSession(taskID, minutes int, note string) (Session, error) {
	if minutes < 1 || minutes > 720 {
		return Session{}, ErrInvalidMins
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var taskTitle string
	for _, task := range s.data.Tasks {
		if task.ID == taskID {
			taskTitle = task.Title
			break
		}
	}
	if taskTitle == "" {
		return Session{}, ErrTaskNotFound
	}
	session := Session{ID: nextID(s.data.Sessions), TaskID: taskID, TaskTitle: taskTitle, Minutes: minutes, Note: note, StudiedAt: time.Now()}
	s.data.Sessions = append(s.data.Sessions, session)
	return session, s.save()
}

func (s *Store) Snapshot() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot := Snapshot{
		Tasks:    append([]Task(nil), s.data.Tasks...),
		Sessions: append([]Session(nil), s.data.Sessions...),
	}
	taskTitles := make(map[int]string, len(snapshot.Tasks))
	for _, task := range snapshot.Tasks {
		taskTitles[task.ID] = task.Title
	}
	for index := range snapshot.Sessions {
		if title := taskTitles[snapshot.Sessions[index].TaskID]; title != "" {
			snapshot.Sessions[index].TaskTitle = title
		}
		if snapshot.Sessions[index].TaskTitle == "" {
			snapshot.Sessions[index].TaskTitle = "Без названия"
		}
	}
	for _, task := range snapshot.Tasks {
		if task.Completed {
			snapshot.CompletedTasks++
		}
	}
	for _, session := range snapshot.Sessions {
		snapshot.TotalMinutes += session.Minutes
	}
	snapshot.Streak = calculateStreak(snapshot.Sessions)
	sort.Slice(snapshot.Sessions, func(i, j int) bool {
		return snapshot.Sessions[i].StudiedAt.After(snapshot.Sessions[j].StudiedAt)
	})
	return snapshot, nil
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, content, 0644)
}

func nextID[T any](items []T) int {
	return len(items) + 1
}

func calculateStreak(sessions []Session) int {
	days := make(map[string]bool)
	for _, session := range sessions {
		days[session.StudiedAt.Local().Format("2006-01-02")] = true
	}
	streak := 0
	today := time.Now()
	for days[today.AddDate(0, 0, -streak).Format("2006-01-02")] {
		streak++
	}
	return streak
}
