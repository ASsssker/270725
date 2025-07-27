package config

import (
	"fmt"
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
	LogLevel string `env:"LOG_LEVEL" envDefault:"info" validate:"oneof=debug info warn error"`
	ServerConfig
	TaskConfig
}

type ServerConfig struct {
	Host         string        `env:"HOST" envDefault:"127.0.0.1"`
	Port         string        `env:"PORT" envDefault:"8080"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"5s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" envDefault:"10"`
}

type TaskConfig struct {
	TasksBufferSize uint `env:"TASKS_BUFFER_SIZE" envDefault:"3"`
	LinksInTask     uint `env:"LINKS_IN_TASK" envDefault:"3"`
}

func MustLoad() Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	return cfg
}
