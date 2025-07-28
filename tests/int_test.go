package tests

import (
	"270725/internal/config"
	"270725/internal/models"
	v1 "270725/internal/rest/v1"
	bp "270725/internal/rest/v1/boileplate"
	"270725/internal/services"
	"270725/internal/storage/inmemory"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

var e = echo.New()

const urlPrefix = "/api/v1"

func TestCreateTaskSuccessfully(t *testing.T) {
	h := setupHandler()

	c, res := createResponser(http.MethodPost, urlPrefix+"/task", "")
	err := h.AddTask(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.Code)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	TaskInfo := &bp.Task{}
	err = json.Unmarshal(body, TaskInfo)
	require.NoError(t, err)
	require.NotEmpty(t, TaskInfo.Id)

	//Ссылки на внешние ресурсы, плохо
	testLinks := []map[string]string{
		{"link": "https://upload.wikimedia.org/wikipedia/commons/4/41/Sunflower_from_Silesia2.jpg"},
		{"link": "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3b/Europeian_diet_Sprite_bottle.jpg/800px-Europeian_diet_Sprite_bottle.jpg"},
		{"link": "https://edu.anarcho-copy.org/Programming%20Languages/Go/build-web-application-with-golang-en.pdf"},
	}

	links, err := json.Marshal(testLinks)
	require.NoError(t, err)

	c, res = createResponser(http.MethodPost, urlPrefix+"/task/"+TaskInfo.Id, string(links))

	err = h.AddLink(c, TaskInfo.Id)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.Code)

	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	// Bad
	time.Sleep(10 * time.Second)

	c, res = createResponser(http.MethodGet, urlPrefix+"/task/"+TaskInfo.Id, "")
	err = h.GetTask(c, TaskInfo.Id)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)

	task := &bp.Task{}
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &task)
	require.NoError(t, err)
	require.Equal(t, TaskInfo.Id, task.Id)
	for idx := range task.FilesLink {
		require.Equal(t, testLinks[idx]["link"], task.FilesLink[idx].Link)
		require.Equal(t, string(models.CompletedTaskLinkStatus), string(task.FilesLink[idx].Status))

	}
	c, res = createResponser(http.MethodGet, urlPrefix+"/task/"+TaskInfo.Id+"/result", "")
	err = h.GetResult(c, TaskInfo.Id)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "application/zip", res.Header().Get(echo.HeaderContentType))

}

func createResponser(method, url string, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c, rec
}

func setupHandler() *v1.Handler {
	cfg := config.MustLoad()
	cfg.ArchivesDir = "./test_archives"

	logger := setupTestLogger()
	repo := inmemory.NewMemory()

	requester := services.NewRequesterService(int(cfg.TasksBufferSize * cfg.LinksInTask))
	archiver, err := services.NewZipper(cfg.ArchivesDir)
	if err != nil {
		panic(fmt.Errorf("failed to create archiver: %w", err))
	}

	taskService := services.NewTaskService(cfg, logger, repo, requester, archiver)

	handler := v1.NewHandler(logger, taskService)
	v1.RegisterHandler(e, handler)

	return handler
}

func setupTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}
