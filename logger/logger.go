package logger

import (
	"fmt"

	"github.com/posener/flarm/flarmport"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	// Database dialect. 'mysql' or 'postgres'.
	Dialect string
	// Database connection string.
	// For mysql: user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	// For postgres: host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai
	URL string
	// Minimal speed to log to DB, in m/s.
	MinLogSpeed int64
}

type Logger struct {
	db  *gorm.DB
	cfg Config
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
		return nil, fmt.Errorf("failed connecting to db: %s", err)
	}
	err = db.AutoMigrate(flarmport.Data{})
	if err != nil {
		return nil, fmt.Errorf("failed migrating table: %s", err)
	}
	return &Logger{db: db, cfg: cfg}, nil
}

func (l *Logger) Log(o flarmport.Data) {
	if l == nil {
		return
	}
	if o.GroundSpeed < l.cfg.MinLogSpeed {
		return
	}
	l.db.Create(o)
}
