package services

import (
	"270725/internal/config"
	"270725/internal/models"
	"270725/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log/slog"
)

type TaskRepository interface {
	NewTask(ctx context.Context) (string, error)
	GetAllTasks(ctx context.Context) ([]*models.Task, error)
	GetTask(ctx context.Context, id string) (*models.Task, error)
	AddLinksToTask(ctx context.Context, taskID string, links []*models.FileLink) (*models.Task, error)
}

type TaskService struct {
	log       *slog.Logger
	taskRepo  TaskRepository
	maxTasks  uint
	validator *validator.Validate
}

func NewTaskService(cfg config.Config, log *slog.Logger, taskRepository TaskRepository) *TaskService {
	return &TaskService{
		log:       log,
		taskRepo:  taskRepository,
		maxTasks:  cfg.TasksBufferSize,
		validator: validator.New(),
	}
}

func (t *TaskService) NewTask(ctx context.Context) (string, error) {
	const op = "taskService.NewTask"
	log := t.log.With(slog.String("op", op))
	log.Debug("start operation")

	taskID, err := t.taskRepo.NewTask(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to add new task: %w", err)
	}

	log.Debug("operation completed")

	return taskID, nil
}

func (t *TaskService) GetAllTasks(ctx context.Context) ([]*models.Task, error) {
	const op = "taskService.GetAllTasks"
	log := t.log.With(slog.String("op", op))
	log.Debug("start operation")

	tasks, err := t.taskRepo.GetAllTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tasks: %w", err)
	}

	log.Debug("operation completed")

	return tasks, nil
}

func (t *TaskService) GetTask(ctx context.Context, id string) (*models.Task, error) {
	const op = "taskService.GetTask"
	log := t.log.With(slog.String("op", op))
	log.Debug("start operation")

	task, err := t.taskRepo.GetTask(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}

		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	log.Debug("operation completed")

	return task, nil
}

func (t *TaskService) AddLinksToTask(ctx context.Context, taskID string, links []*models.FileLink) (*models.Task, error) {
	const op = "taskService.AddLinksToTask"
	log := t.log.With(slog.String("op", op))
	log.Debug("start operation")

	task, err := t.taskRepo.GetTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}

		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if len(links)+len(task.FilesLink) > int(t.maxTasks) {
		return nil, fmt.Errorf("max tasks reached: %w", ErrValidation)
	}

	if err := t.validator.Struct(task); err != nil {
		return nil, fmt.Errorf("failed to validate task links: %w", err)
	}

	// TODO: if links count eq max, send start proccess

	log.Debug("operation completed")

	return nil, nil
}
