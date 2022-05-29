package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func (s *server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	resp := make(map[string]string)

	switch r.Method {
	case http.MethodPost:

		u, err := url.Parse(r.RequestURI)

		if err != nil {
			log.Printf("parse request uri error: %v", err)
			resp["error"] = "Incorrect URI"
			w.WriteHeader(http.StatusBadRequest)
			break
		}

		pathSlice := strings.Split(u.Path, "/")

		if len(pathSlice) != 5 {
			resp["error"] = "Incorrect URI"
			w.WriteHeader(http.StatusNotFound)
			break
		}

		metricType, metricName, metricValue := pathSlice[2], pathSlice[3], pathSlice[4]

		err = s.storage.Save(metricType, metricName, metricValue)

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
			break
		}

		resp["message"] = "Success save metric"
		w.WriteHeader(http.StatusOK)

	default:
		resp["error"] = fmt.Sprintf("Incorrect http method for URI %s", r.RequestURI)
		w.WriteHeader(405)
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
