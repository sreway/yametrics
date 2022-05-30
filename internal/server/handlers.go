package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates/index.gohtml
var templatesFS embed.FS

var templateFiles = map[string]string{
	"/": "templates/index.gohtml",
}

func (s *server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	resp := make(map[string]string)

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	err := s.saveMetric(metricType, metricName, metricValue)

	if err != nil {
		log.Printf("storage save: %v", err)
		resp["error"] = "Can't save metric"
		switch err {
		case ErrInvalidMetricType:
			w.WriteHeader(http.StatusNotImplemented)
		case ErrInvalidMetricValue:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	} else {
		resp["message"] = "Success save metric"
		w.WriteHeader(http.StatusOK)
	}

	jsonResp, err := json.Marshal(resp)

	if err != nil {
		log.Printf("update metric: error creating json response: %v", err)
	}
	_, err = w.Write(jsonResp)

	if err != nil {
		log.Printf("update metric: error write json response: %v", err)
	}
}

func (s *server) Index(w http.ResponseWriter, r *http.Request) {
	templatePattern, ok := templateFiles[r.URL.Path]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metrics := s.getMetrics()

	tmpl, err := template.ParseFS(templatesFS, templatePattern)

	if err != nil {
		log.Printf("parsing template error: %v", err)
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	err = tmpl.Execute(w, metrics)

	if err != nil {
		log.Printf("index error: %v", err)
		w.WriteHeader(http.StatusNotImplemented)
	}

}

func (s *server) MetricValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	metricName := chi.URLParam(r, "metricName")
	metricType := chi.URLParam(r, "metricType")
	val, err := s.getMetricValue(metricType, metricName)

	if err != nil {
		switch err {
		case ErrInvalidMetricValue:
			w.WriteHeader(http.StatusBadRequest)
		case ErrNotFoundMetric:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotImplemented)
			log.Printf("get metric value: %v", err)
		}
	}
	_, err = w.Write([]byte(fmt.Sprintf("%v", val)))

	if err != nil {
		w.WriteHeader(http.StatusNotImplemented)
		log.Printf("get metric value: error write bytes response: %v", err)
	}
}
