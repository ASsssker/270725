package inmemory

import (
	"270725/internal/models"
	"270725/internal/storage"
	"context"
	"github.com/google/uuid"
	"sync"
)

type Memory struct {
	tasks map[string]*models.Task
	mu    *sync.RWMutex
}

func NewMemory() *Memory {
	return &Memory{
		tasks: make(map[string]*models.Task),
		mu:    &sync.RWMutex{},
	}
}

func (m *Memory) NewTask(_ context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task := &models.Task{
		ID: uuid.NewString(),
	}
	m.tasks[task.ID] = task

	return task.ID, nil
}

func (m *Memory) GetAllTasks(_ context.Context) ([]*models.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (m *Memory) GetTask(_ context.Context, id string) (*models.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[id]
	if !exists {
		return nil, storage.ErrTaskNotFound
	}

	return task, nil
}

func (m *Memory) AddLinksToTask(_ context.Context, taskID string, links []*models.FileLink) (*models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, storage.ErrTaskNotFound
	}

	for _, fileLink := range links {
		task.FilesLink = append(task.FilesLink, fileLink)
	}
	m.tasks[taskID] = task

	return task, nil
}
