package server

import "github.com/go-chi/chi/v5"

func (s *server) initRoutes(r *chi.Mux) {
	r.Post("/update/{metricType}/{metricName}/{metricValue}", s.UpdateMetric)
	r.Get("/value/{metricType}/{metricName}", s.MetricValue)
	r.Get("/", s.Index)
}
