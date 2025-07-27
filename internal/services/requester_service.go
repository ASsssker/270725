package services

import (
	"fmt"
	"github.com/alitto/pond/v2"
	"io"
	"log/slog"
	"net/http"
)

type Requester struct {
	client *http.Client
	pool   pond.Pool
}

type responseInfo struct {
	link   string
	result []byte
}

func NewRequesterService(poolSize int) *Requester {
	return &Requester{
		client: &http.Client{},
		pool:   pond.NewPool(poolSize, pond.WithNonBlocking(true)),
	}
}

func (r *Requester) GetLinksContents(log *slog.Logger, links []string) map[string][]byte {
	resultsChan := make(chan responseInfo)
	tasks := make([]pond.Task, 0, len(links))
	for _, link := range links {
		task := r.pool.Submit(func() {
			data, err := r.request(link)
			if err != nil {
				log.Error("failed to send request", slog.String("link", link), slog.String("error", err.Error()))
				return
			}

			resultsChan <- responseInfo{link: link, result: data}
		})

		tasks = append(tasks, task)
	}

	go func() {
		for _, task := range tasks {
			if err := task.Wait(); err != nil {
				log.Error("failed to wait task", slog.String("error", err.Error()))
			}
		}

		close(resultsChan)
	}()

	results := make(map[string][]byte)
	for taskInfo := range resultsChan {
		results[taskInfo.link] = taskInfo.result
	}

	return results

}

func (r *Requester) request(link string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	response, err := r.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
