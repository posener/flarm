package logger

import (
	"fmt"

	"github.com/posener/flarm/process"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Dialect string
	URL     string
}

type Logger struct {
	db *gorm.DB
}

func New(cfg Config) (*Logger, error) {
	if cfg.URL == "" {
		return nil, nil
	}
	var d gorm.Dialector
	switch cfg.Dialect {
	case "postgres":
		d = postgres.Open(cfg.URL)
	case "mysql":
		d = mysql.Open(cfg.URL)
	default:
		return nil, fmt.Errorf("unsupported dialect %q", cfg.Dialect)
	}
	db, err := gorm.Open(d, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(process.Object{})
	return &Logger{db: db}, nil
}

func (l *Logger) Log(o *process.Object) {
	if l == nil {
		return
	}
	l.db.Create(o)
}
