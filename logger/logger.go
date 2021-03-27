package logger

import (
	"encoding/json"
	"log"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Path         string
	RotateSizeMB int
}

type Logger struct {
	l *log.Logger
}

func New(cfg Config) *Logger {
	if cfg.Path == "" {
		return nil
	}
	return &Logger{
		l: log.New(&lumberjack.Logger{
			Filename: cfg.Path,
			MaxSize:  cfg.RotateSizeMB,
		}, "", 0),
	}
}

func (l *Logger) Log(v interface{}) {
	if l == nil {
		return
	}

	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Failed marshaling %+v: %s", v, err)
		return
	}
	l.l.Print(string(data))
}
