// Package storage implements and describes a repository for collecting and storing metrics
package storage

import (
	"context"
	"errors"
	"os"
	"sync"

	//nolint:nolintlint
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"

	"github.com/sreway/yametrics/internal/metrics"
)

var (
	ErrNotFoundMetric     = errors.New("not found metric")
	ErrStoreMetrics       = errors.New("can't store metrics")
	ErrLoadMetrics        = errors.New("can't load metrics")
	ErrStorageUnavailable = errors.New("storage unavailable")
)

type (
	memoryStorage struct {
		metrics metrics.Metrics
		mu      sync.RWMutex
		fileObj *os.File
	}

	pgStorage struct {
		connection *pgx.Conn
	}
	// Storage describes the implementation of storage (repository)
	Storage interface {
		Save(ctx context.Context, metric metrics.Metric) error
		GetMetric(ctx context.Context, metricType, metricID string) (*metrics.Metric, error)
		GetMetrics(ctx context.Context) (*metrics.Metrics, error)
		IncrementCounter(ctx context.Context, metricID string, value int64) error
		BatchMetrics(ctx context.Context, m []metrics.Metric) error
		Close(ctx context.Context) error
	}
	// MemoryStorage describes the implementation of in-memory storage
	MemoryStorage interface {
		Storage
		LoadMetrics() error
		StoreMetrics() error
	}
	// PgStorage describes the implementation of PostgreSQL storage
	PgStorage interface {
		Storage
		Ping(ctx context.Context) error
		ValidateSchema(sourceMigrationsURL string) error
	}
)
