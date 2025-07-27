package models

type TaskLinkStatus string

const (
	NewTaskLinkStatus       TaskLinkStatus = "new"
	InProcessTaskLinkStatus TaskLinkStatus = "in_process"
	CompletedTaskLinkStatus TaskLinkStatus = "completed"
	ErrorTaskLinkStatus     TaskLinkStatus = "error"
)

type Task struct {
	ID        string
	FilesLink []*FileLink
}

type FileLink struct {
	Link   string `validate:"url"`
	Status TaskLinkStatus
}
