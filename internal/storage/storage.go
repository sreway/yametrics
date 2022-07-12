package storage

import (
	"context"
	"errors"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	"github.com/sreway/yametrics/internal/metrics"
	"os"
	"sync"
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

	Storage interface {
		Save(ctx context.Context, metric metrics.Metric) error
		GetMetric(ctx context.Context, metricType, metricID string) (*metrics.Metric, error)
		GetMetrics(ctx context.Context) (*metrics.Metrics, error)
		IncrementCounter(ctx context.Context, metricID string, value int64) error
		BatchMetrics(ctx context.Context, m []metrics.Metric) error
		Close() error
	}

	MemoryStorage interface {
		Storage
		LoadMetrics() error
		StoreMetrics() error
	}

	PgStorage interface {
		Storage
		Ping(ctx context.Context) error
		ValidateSchema(sourceMigrationsURL string) error
	}
)
