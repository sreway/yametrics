package server

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_server_UpdateMetric(t *testing.T) {

	type want struct {
		statusCode int
	}

	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "send counter",
			request: "/update/counter/PollCount/100",
			want: want{
				statusCode: 200,
			},
		},

		{
			name:    "send gauge",
			request: "/update/gauge/RandomValue/10.8",
			want: want{
				statusCode: 200,
			},
		},

		{
			name:    "invalid value",
			request: "/update/counter/PollCount/none",
			want: want{
				statusCode: 400,
			},
		},

		{
			name:    "invalid type",
			request: "/update/unknown/PollCount/100",
			want: want{
				statusCode: 501,
			},
		},

		{
			name:    "invalid uri",
			request: "/update/unknown",
			want: want{
				statusCode: 404,
			},
		},
	}

	s := &server{
		&http.Server{},
		NewStorage(),
		nil,
		nil,
		nil,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.UpdateMetric)
			h.ServeHTTP(w, request)
			res := w.Result()

			_, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			err = res.Body.Close()

			if err != nil {
				t.Fatal(err)
			}
		})
	}

}
