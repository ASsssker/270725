package services

import (
	"270725/internal/config"
	"270725/internal/models"
	"270725/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/alitto/pond/v2"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"path/filepath"
	"strings"
	"sync/atomic"
)

type TaskRepository interface {
	NewTask(ctx context.Context) (string, error)
	GetAllTasks(ctx context.Context) ([]*models.Task, error)
	GetTask(ctx context.Context, id string) (*models.Task, error)
	AddLinksToTask(ctx context.Context, taskID string, links []*models.FileLink) (*models.Task, error)
	MarkTaskLinksInProcessStatus(ctx context.Context, taskID string) error
	MarkTaskLinksCompleted(ctx context.Context, taskID string, completedLinks []string) error
}

type RequesterClient interface {
	GetLinksContents(log *slog.Logger, links []string) map[string][]byte
}

type Archiver interface {
	ToArchive(archiveName string, files map[string][]byte) error
}

type TaskService struct {
	log               *slog.Logger
	taskRepo          TaskRepository
	requester         RequesterClient
	archiver          Archiver
	taskInProcess     atomic.Int64
	maxTasks          uint
	linksInFile       uint
	validator         *validator.Validate
	pool              pond.Pool
	allowedExtensions []string
}

func NewTaskService(
	cfg config.Config,
	log *slog.Logger,
	taskRepository TaskRepository,
	requester RequesterClient,
	archiver Archiver,
) *TaskService {
	return &TaskService{
		log:               log,
		taskRepo:          taskRepository,
		requester:         requester,
		archiver:          archiver,
		taskInProcess:     atomic.Int64{},
		maxTasks:          cfg.TasksBufferSize,
		linksInFile:       cfg.LinksInTask,
		validator:         validator.New(),
		pool:              pond.NewPool(int(cfg.TasksBufferSize), pond.WithNonBlocking(true)),
		allowedExtensions: cfg.AllowedExtensions,
	}
}

func (t *TaskService) NewTask(ctx context.Context) (string, error) {
	const op = "taskService.NewTask"
	log := t.log.With(slog.String("op", op))
	log.Debug("start operation")

	if t.taskInProcess.Load() > int64(t.maxTasks) {
		return "", ErrServiceBusy
	}

	taskID, err := t.taskRepo.NewTask(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to add new task: %w", err)
	}
	t.taskInProcess.Add(1)

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

	if len(links)+len(task.FilesLink) > int(t.linksInFile) {
		return nil, fmt.Errorf("max tasks reached: %w", ErrValidation)
	}

	if err := t.validator.Struct(task); err != nil {
		return nil, fmt.Errorf("failed to validate task links: %w", err)
	}

	if err := t.checkLinksExtension(links); err != nil {
		return nil, fmt.Errorf("failed to check extensions: %w: %w", err, ErrValidation)
	}

	for _, fileLink := range links {
		fileLink.Status = models.NewTaskLinkStatus
	}
	task, err = t.taskRepo.AddLinksToTask(ctx, taskID, links)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}

		return nil, fmt.Errorf("failed to add links to task: %w", err)
	}

	if len(task.FilesLink) == int(t.linksInFile) {
		_, ok := t.pool.TrySubmit(func() {
			t.processTask(task)
		})
		if !ok {
			return nil, ErrServiceBusy
		}
	}

	log.Debug("operation completed")

	return task, nil
}

func (t *TaskService) processTask(task *models.Task) {
	t.pool.Go(func() {
		defer t.taskInProcess.Add(-1)
		if err := t.taskRepo.MarkTaskLinksInProcessStatus(context.TODO(), task.ID); err != nil {
			t.log.Error("failed to update task status to in process", slog.String("error", err.Error()))

			if err := t.taskRepo.MarkTaskLinksCompleted(context.TODO(), task.ID, []string{}); err != nil {
				t.log.Error("failed to update task status to error", slog.String("error", err.Error()))
			}

			return
		}

		log := t.log.With(slog.String("task_id", task.ID))
		linkContents := t.requester.GetLinksContents(log, getLinksFromTask(task))

		if err := t.archiver.ToArchive(task.ID, convertLinksFilename(linkContents)); err != nil {
			t.log.Error("failed to archive task", slog.String("error", err.Error()))
			if err := t.taskRepo.MarkTaskLinksCompleted(context.TODO(), task.ID, []string{}); err != nil {
				t.log.Error("failed to update task status to error", slog.String("error", err.Error()))
			}

			return
		}

		if err := t.taskRepo.MarkTaskLinksCompleted(context.TODO(), task.ID, getLinksFromMap(linkContents)); err != nil {
			t.log.Error("failed to update task status to completed", slog.String("error", err.Error()))
			return
		}
	})
}

func (t *TaskService) checkLinksExtension(links []*models.FileLink) error {
	for _, link := range links {
		allowed := false
		for _, extension := range t.allowedExtensions {
			if strings.HasSuffix(link.Link, extension) {
				allowed = true
			}
		}
		if !allowed {
			return fmt.Errorf(`link "%s" extension not allowed, allowed extensions %s`, link.Link, strings.Join(t.allowedExtensions, ","))

		}
	}
	return nil
}

func convertLinksFilename(linksInfo map[string][]byte) map[string][]byte {
	result := make(map[string][]byte)
	for link, data := range linksInfo {
		_, name := filepath.Split(link)
		result[name] = data
	}

	return result
}

func getLinksFromTask(task *models.Task) []string {
	links := make([]string, 0, len(task.FilesLink))
	for _, link := range task.FilesLink {
		links = append(links, link.Link)
	}

	return links
}

func getLinksFromMap(linksInfo map[string][]byte) []string {
	links := make([]string, 0, len(linksInfo))
	for key := range linksInfo {
		links = append(links, key)
	}

	return links
}
