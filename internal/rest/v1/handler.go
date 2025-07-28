package v1

import (
	"270725/internal/models"
	bp "270725/internal/rest/v1/boileplate"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log/slog"
	"net/http"
)

type TaskService interface {
	NewTask(ctx context.Context) (string, error)
	GetAllTasks(ctx context.Context) ([]*models.Task, error)
	GetTask(ctx context.Context, id string) (*models.Task, error)
	AddLinksToTask(ctx context.Context, taskID string, links []*models.FileLink) (*models.Task, error)
	GetTaskResult(ctx context.Context, taskID string) (string, string, error)
}

func RegisterHandler(router *echo.Echo, handler *Handler) {
	router.Use(middleware.Recover(), handler.handleError())
	bp.RegisterHandlersWithBaseURL(router, handler, "/api/v1")
}

type Handler struct {
	log         *slog.Logger
	taskService TaskService
}

func NewHandler(log *slog.Logger, taskService TaskService) *Handler {
	return &Handler{
		log:         log,
		taskService: taskService,
	}
}

func (h *Handler) AddTask(c echo.Context) error {
	ctx := c.Request().Context()

	taskID, err := h.taskService.NewTask(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new task: %w", err)
	}

	return c.JSON(http.StatusCreated, bp.Task{Id: taskID})
}

func (h *Handler) GetAllTasks(c echo.Context) error {
	ctx := c.Request().Context()

	tasks, err := h.taskService.GetAllTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all tasks: %w", err)
	}

	tasksResponse := make([]bp.Task, 0, len(tasks))
	for _, task := range tasks {
		tasksResponse = append(tasksResponse, bp.Task{
			Id:        task.ID,
			FilesLink: convertLinks(task.FilesLink),
		})
	}

	return c.JSON(http.StatusOK, tasksResponse)
}

func (h *Handler) AddLink(c echo.Context, id string) error {
	ctx := c.Request().Context()

	links := bp.AddLinkJSONRequestBody{}
	if err := c.Bind(&links); err != nil {
		return fmt.Errorf("failed to bind add link: %w", err)
	}

	linksModels := convertRequestLink(links)
	task, err := h.taskService.AddLinksToTask(ctx, id, linksModels)
	if err != nil {
		return fmt.Errorf("failed to add link: %w", err)
	}

	responseTask := bp.Task{
		Id:        task.ID,
		FilesLink: convertLinks(task.FilesLink),
	}

	return c.JSON(http.StatusCreated, responseTask)
}

func (h *Handler) GetTask(c echo.Context, id string) error {
	ctx := c.Request().Context()

	task, err := h.taskService.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	responseTask := bp.Task{
		Id:        task.ID,
		FilesLink: convertLinks(task.FilesLink),
	}

	return c.JSON(http.StatusOK, responseTask)
}

func (h *Handler) GetResult(c echo.Context, id string) error {
	ctx := c.Request().Context()

	filePath, name, err := h.taskService.GetTaskResult(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get result: %w", err)
	}

	return c.Attachment(filePath, name)
}

func convertLinks(links []*models.FileLink) []bp.FileLinkInfo {
	fileLinksInfo := make([]bp.FileLinkInfo, 0, len(links))
	for _, link := range links {
		linkInfo := bp.FileLinkInfo{
			Link:   link.Link,
			Status: convertLinkStatus(link.Status),
		}

		fileLinksInfo = append(fileLinksInfo, linkInfo)
	}

	return fileLinksInfo
}

func convertLinkStatus(status models.TaskLinkStatus) bp.FileLinkInfoStatus {
	switch status {
	case models.NewTaskLinkStatus:
		return bp.FileLinkInfoStatusNew
	case models.InProcessTaskLinkStatus:
		return bp.FileLinkInfoStatusInProcess
	case models.CompletedTaskLinkStatus:
		return bp.FileLinkInfoStatusCompleted
	case models.ErrorTaskLinkStatus:
		return bp.FileLinkInfoStatusError
	}

	panic("invalid task link status")
}

func convertRequestLink(links bp.AddLinkJSONRequestBody) []*models.FileLink {
	fileLinksInfo := make([]*models.FileLink, 0, len(links))
	for _, link := range links {
		fileLinksInfo = append(fileLinksInfo, &models.FileLink{
			Link: link.Link,
		})
	}

	return fileLinksInfo
}
