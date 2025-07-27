package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelError LogLevel = "error"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
)

type Config struct {
	LogLevel LogLevel `env:"LOG_LEVEL" env-default:"info" validate:"oneof=debug info warn error"`
	ServerConfig
	TaskConfig
	Filter
}

type ServerConfig struct {
	Host         string        `env:"HOST" env-default:"127.0.0.1"`
	Port         string        `env:"PORT" env-default:"8080"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" env-default:"5s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" env-default:"10s"`
}

type TaskConfig struct {
	TasksBufferSize uint `env:"TASKS_BUFFER_SIZE" env-default:"3"`
	LinksInTask     uint `env:"LINKS_IN_TASK" env-default:"3"`
}

type Filter struct {
	AllowedExtensions []string `env:"ALLOWED_EXTENSIONS" env-default:"jpg,png,pdf"`
}

func MustLoad() Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		panic(fmt.Errorf("failed to validate config: %w", err))
	}

	return cfg
}
