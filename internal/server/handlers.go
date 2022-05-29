package server

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
)

var indexTmpl = `
<!DOCTYPE html>
<html>
    <head>
    <title>YaMetrics</title>
    </head>
    <body>
		{{range $mtype, $metrics := .}}
		{{range $mname, $mvalue := $metrics}}
		<p>{{$mname}}: {{$mvalue}}</p>
		{{end}}
		{{end}}
    </body>
</html>`

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

	metrics := s.getMetrics()
	tmpl := template.Must(template.New("metrics").Parse(indexTmpl))
	err := tmpl.Execute(w, metrics)

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
