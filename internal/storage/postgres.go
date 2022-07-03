package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/sreway/yametrics/internal/metrics"
	"log"
)

func NewPgStorage(ctx context.Context, dsn string) (PgStorage, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("NewPgStroge error: %w", err)
	}

	log.Println("NewPgStorage: success connect database")

	return &pgStorage{
		connection: conn,
	}, nil
}

func (s *pgStorage) Save(ctx context.Context, metric metrics.Metric) error {
	_, err := s.connection.Exec(ctx, "INSERT INTO metrics (name, type, delta, value) VALUES ($1, $2, $3, $4)"+
		"ON CONFLICT ON CONSTRAINT uniq_name_type DO UPDATE set delta=$3, value=$4",
		metric.ID, metric.MType, metric.Int64Pointer(), metric.Float64Pointer())
	if err != nil {
		return fmt.Errorf("pgStorage_Save:%w", err)
	}

	return nil
}

func (s *pgStorage) GetMetric(ctx context.Context, metricType, metricID string) (*metrics.Metric, error) {
	var (
		m metrics.Metric
	)

	q := fmt.Sprintf("SELECT delta, value FROM metrics WHERE name = '%s' and type = '%s'", metricID, metricType)
	err := s.connection.QueryRow(ctx, q).Scan(&m.Delta, &m.Value)

	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, fmt.Errorf("pgStorage_GetMetric: %w", ErrNotFoundMetric)
		case errors.As(err, &pgErr):
			switch pgErr.Code {
			case "42703":
				return nil, fmt.Errorf("pgStorage_GetMetric: %w", ErrNotFoundMetric)
			}
		default:
			return nil, fmt.Errorf("pgStorage_GetMetric: %w", err)
		}
	}

	m.ID = metricID
	m.MType = metricType
	return &m, nil
}

func (s *pgStorage) GetMetrics(ctx context.Context) (*metrics.Metrics, error) {

	m := metrics.Metrics{
		Counter: make(map[string]metrics.Metric),
		Gauge:   make(map[string]metrics.Metric),
	}

	rows, err := s.connection.Query(ctx, "SELECT name, type, delta, value FROM metrics")

	if err != nil {
		return nil, fmt.Errorf("Server_GetMetrics: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var v metrics.Metric
		err = rows.Scan(&v.ID, &v.MType, &v.Delta, &v.Value)

		if err != nil {
			return nil, fmt.Errorf("Server_GetMetrics: %w", err)
		}

		switch v.MType {
		case "counter":
			m.Counter[v.ID] = v
		case "gauge":
			m.Gauge[v.ID] = v
		default:
			return nil, fmt.Errorf("Server_GetMetrics: %w", metrics.ErrInvalidMetricType)
		}
	}

	err = rows.Err()

	if err != nil {
		return nil, fmt.Errorf("Server_GetMetrics: %w", err)
	}

	return &m, nil
}

func (s *pgStorage) IncrementCounter(ctx context.Context, metricID string, value int64) error {
	_, err := s.connection.Exec(ctx, "INSERT INTO metrics (name, type, delta) VALUES ($1, $2, $3)"+
		"ON CONFLICT ON CONSTRAINT uniq_name_type DO UPDATE set delta = $3 + metrics.delta", metricID, "counter", value)
	if err != nil {
		return fmt.Errorf("pgStorage_Save:%w", err)
	}

	return nil

}

func (s *pgStorage) Ping(ctx context.Context) error {
	if err := s.connection.Ping(ctx); err != nil {
		return fmt.Errorf("pgStorage_Ping: %w", ErrStorageUnavailable)
	}
	return nil
}

// Нужно ли закрывать https://pkg.go.dev/database/sql#Open ?
func (s *pgStorage) Close() error {
	if err := s.connection.Close(context.Background()); err != nil {
		return fmt.Errorf("pgStorage_Close: %w", err)
	}
	return nil
}

func (s *pgStorage) ValidateSchema(sourceMigrationsURL string) error {
	config := s.connection.Config()
	migrateURL := fmt.Sprintf("pgx://%s:%s@%s:%d/%s",
		config.User, config.Password, config.Host, config.Port, config.Database)
	m, err := migrate.New(sourceMigrationsURL, migrateURL)

	if err != nil {
		return fmt.Errorf("pgStorage_ValidateSchema: %w", err)
	}

	err = m.Up()

	if err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			log.Printf("Migrate up: %s", err)
		default:
			return fmt.Errorf("pgStorage_ValidateSchema: %w", err)
		}
	}
	return nil
}

func (s *pgStorage) BatchMetrics(ctx context.Context, m []metrics.Metric) error {
	return nil
}
