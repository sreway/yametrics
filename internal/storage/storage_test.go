package storage

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sreway/yametrics/internal/metrics"
)

func NewTestMemoryStorage(metricID, metricType, metricValue, storageFile string) (MemoryStorage, error) {
	metric, err := metrics.NewMetric(metricID, metricType, metricValue)
	if err != nil {
		return nil, err
	}

	testStorage, err := NewMemoryStorage(storageFile)
	if err != nil {
		return nil, err
	}

	err = testStorage.Save(context.Background(), metric)

	if err != nil {
		return nil, err
	}

	return testStorage, err
}

func OpenTestFile(path string) (*os.File, error) {
	flag := os.O_RDWR | os.O_CREATE
	fileObj, err := os.OpenFile(path, flag, 0o644)
	if err != nil {
		return nil, fmt.Errorf("NewFileObj: can't open file %s", path)
	}
	return fileObj, nil
}

func Test_storage_Save(t *testing.T) {
	type args struct {
		metricID    string
		metricType  string
		metricValue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "save counter",
			args: args{
				metricType:  "counter",
				metricID:    "PollCount",
				metricValue: "10",
			},
			wantErr: false,
		},

		{
			name: "save gauge",
			args: args{
				metricType:  "gauge",
				metricID:    "RandomValue",
				metricValue: "1.1",
			},
			wantErr: false,
		},

		{
			name: "invalid type",
			args: args{
				metricType:  "invalid",
				metricID:    "RandomValue",
				metricValue: "1.1",
			},
			wantErr: true,
		},
	}
	s, err := NewMemoryStorage("")
	assert.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, _ := metrics.NewMetric(tt.args.metricID, tt.args.metricType, tt.args.metricValue)
			if err := s.Save(context.Background(), metric); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_storage_GetMetric(t *testing.T) {
	type storageData struct {
		metricID    string
		metricType  string
		metricValue string
	}

	type fields struct {
		storageData storageData
	}

	type args struct {
		metricID   string
		metricType string
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "get counter",
			args: args{
				metricType: "counter",
				metricID:   "PollCount",
			},
			fields: fields{
				storageData: storageData{
					metricID:    "PollCount",
					metricType:  "counter",
					metricValue: "10",
				},
			},
			wantErr: false,
		},

		{
			name: "get gauge",
			args: args{
				metricType: "gauge",
				metricID:   "testGauge",
			},
			fields: fields{
				storageData: storageData{
					metricID:    "testGauge",
					metricType:  "gauge",
					metricValue: "10.1",
				},
			},
			wantErr: false,
		},

		{
			name: "invalid type",
			args: args{
				metricType: "invalid",
				metricID:   "RandomValue",
			},
			wantErr: true,
		},

		{
			name: "non existent",
			args: args{
				metricType: "counter",
				metricID:   "RandomValue",
			},
			wantErr: true,
		},
	}
	var m Storage

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.storageData != (storageData{}) {
				testStorage, err := NewTestMemoryStorage(tt.fields.storageData.metricID,
					tt.fields.storageData.metricType, tt.fields.storageData.metricValue, "")
				assert.NoError(t, err)
				m = testStorage
			} else {
				memStorage, err := NewMemoryStorage("")
				assert.NoError(t, err)
				m = memStorage
			}

			if _, err := m.GetMetric(context.Background(), tt.args.metricType, tt.args.metricID); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_storage_StoreMetrics(t *testing.T) {
	type storageData struct {
		metricID    string
		metricType  string
		metricValue string
	}

	type fields struct {
		storageData storageData
	}

	type args struct {
		filePath string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "save metrics",
			fields: fields{
				storageData: storageData{
					metricID:    "testCounter",
					metricType:  "counter",
					metricValue: "10",
				},
			},

			args: args{
				filePath: "/tmp/test.json",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewTestMemoryStorage(tt.fields.storageData.metricID, tt.fields.storageData.metricType,
				tt.fields.storageData.metricValue, tt.args.filePath)
			assert.NoError(t, err)
			fileObj, err := OpenTestFile(tt.args.filePath)
			defer func() {
				err = fileObj.Close()
				assert.NoError(t, err)
			}()
			assert.NoError(t, err)
			tt.wantErr(t, s.StoreMetrics(), fmt.Sprintf("StoreMetrics(%v)", tt.args.filePath))
			defer os.Remove(tt.args.filePath)
		})
	}
}

func Test_storage_LoadMetrics(t *testing.T) {
	type storageData struct {
		metricID    string
		metricType  string
		metricValue string
	}

	type fields struct {
		storageData storageData
	}

	type args struct {
		filePath string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "load metrics",
			fields: fields{
				storageData: storageData{
					metricID:    "testCounter",
					metricType:  "counter",
					metricValue: "10",
				},
			},

			args: args{
				filePath: "/tmp/test.json",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s, err := NewTestMemoryStorage(tt.fields.storageData.metricID, tt.fields.storageData.metricType,
				tt.fields.storageData.metricValue, tt.args.filePath)
			assert.NoError(t, err)
			err = s.StoreMetrics()
			assert.NoError(t, err)
			err = s.Close(ctx)
			assert.NoError(t, err)
			emptyStore, err := NewMemoryStorage(tt.args.filePath)
			assert.NoError(t, err)
			tt.wantErr(t, emptyStore.LoadMetrics(), fmt.Sprintf("LoadMetrics(%v)", tt.args.filePath))
			defer os.Remove(tt.args.filePath)
			defer emptyStore.Close(ctx)
		})
	}
}
