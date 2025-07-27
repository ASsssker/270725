package services

import "errors"

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrValidation   = errors.New("validation error")
)
