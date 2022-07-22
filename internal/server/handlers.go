package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sreway/yametrics/internal/metrics"
	"github.com/sreway/yametrics/internal/storage"
)

var (
	//go:embed templates/index.gohtml
	templatesFS   embed.FS
	templateFiles = map[string]string{
		"/": "templates/index.gohtml",
	}
)

func (s *server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	m, err := metrics.NewMetric(metricName, metricType, metricValue)
	if err != nil {
		log.Printf("Server_UpdateMetric: %s", err.Error())
		ErrHandel(w, err)
		return
	}

	err = s.saveMetric(r.Context(), m, false)

	if err != nil {
		log.Printf("Server_UpdateMetric: %s", err.Error())
		ErrHandel(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	templatePattern, ok := templateFiles[r.URL.Path]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sMetrics, err := s.getMetrics(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	tmpl, err := template.ParseFS(templatesFS, templatePattern)
	if err != nil {
		log.Printf("Server_Index: parsing template error: %v", err)
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	err = tmpl.Execute(w, sMetrics)
	if err != nil {
		log.Printf("index error: %v", err)
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func (s *server) MetricValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	metricName := chi.URLParam(r, "metricName")
	metricType := chi.URLParam(r, "metricType")

	metric, err := s.getMetric(r.Context(), metricType, metricName, false)
	if err != nil {
		log.Printf("Server_MetricValue: %s", err.Error())
		ErrHandel(w, err)
		return
	}

	_, err = w.Write([]byte(metric.GetStrValue()))
	if err != nil {
		w.WriteHeader(http.StatusNotImplemented)
		log.Printf("Server_MetricValue: get metric value: error write bytes response: %v", err)
	}
}

func (s *server) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	var m metrics.Metric
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&m); err != nil {
		log.Printf("Server_UpdateMetricJSON: can't decode body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := s.saveMetric(r.Context(), m, s.cfg.Key != "")
	if err != nil {
		log.Printf("Server_UpdateMetricJSON: %s", err.Error())
		ErrHandel(w, err)
		return
	}

	storageMetric, err := s.getMetric(r.Context(), m.MType, m.ID, s.cfg.Key != "")
	if err != nil {
		log.Printf("Server_UpdateMetricJSON: %s", err.Error())
		ErrHandel(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(&storageMetric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("failed encode metric: %v", err)
		return
	}
}

func (s *server) MetricValueJSON(w http.ResponseWriter, r *http.Request) {
	var m metrics.Metric
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&m); err != nil {
		log.Printf("Server_MetricValueJSON: can't decode body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sMetric, err := s.getMetric(r.Context(), m.MType, m.ID, s.cfg.Key != "")
	if err != nil {
		log.Printf("Server_MetricValueJSON: %s", err.Error())
		ErrHandel(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(&sMetric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Server_MetricValueJSON: failed encode metric: %v", err)
		return
	}
}

func (s *server) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	err := s.pingStorage(ctx)
	if err != nil {
		log.Printf("Server_Ping: %s", err.Error())
		ErrHandel(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) BatchMetrics(w http.ResponseWriter, r *http.Request) {
	var m []metrics.Metric
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&m); err != nil {
		log.Printf("Server_BatchMetrics: can't decode body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := s.batchMetrics(r.Context(), m, s.cfg.Key != "")
	if err != nil {
		log.Printf("Server_BatchMetrics: %s", err.Error())
		ErrHandel(w, err)
		return
	}

	storageMetrics, err := s.getMetricsList(r.Context(), s.cfg.Key != "")
	if err != nil {
		log.Printf("Server_BatchMetrics: %s", err.Error())
		ErrHandel(w, err)
		return
	}

	var stdout struct {
		Metrics []metrics.Metric
	}
	stdout.Metrics = storageMetrics
	if err := json.NewEncoder(w).Encode(&stdout); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("failed encode metric: %v", err)
		return
	}
}

func ErrHandel(w http.ResponseWriter, err error) {
	var metricErr *metrics.ErrMetric

	if err == nil {
		return
	}

	if errors.As(err, &metricErr) {
		switch {
		case errors.Is(metricErr.MetricError, metrics.ErrInvalidMetricType):
			w.WriteHeader(http.StatusNotImplemented)
		case errors.Is(metricErr.MetricError, metrics.ErrInvalidMetricValue):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(metricErr.MetricError, ErrInvalidMetricHash):
			w.WriteHeader(http.StatusBadRequest)
		default:

			w.WriteHeader(http.StatusNotImplemented)
		}
		return
	}

	switch {
	case errors.Is(err, storage.ErrNotFoundMetric):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, storage.ErrStorageUnavailable):
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}
