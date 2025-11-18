package metrics

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/rs/zerolog/log"
)

type MemRepository struct {
	mx      *sync.Mutex
	metrics map[string]any
}

func NewMemRepository() MemRepository {
	return MemRepository{mx: &sync.Mutex{}, metrics: make(map[string]any)}
}

func (s MemRepository) UpdateMetric(name string, value any) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	if name == "" {
		return errors.New("empty metric name")
	}
	if value == nil {
		return errors.New("empty metric value")
	}
	switch reflect.ValueOf(value).Kind() {
	case reflect.Float64:
		s.metrics[name] = value
	case reflect.Uint64:
		s.metrics[name] = float64(value.(uint64))
	case reflect.Uint32:
		s.metrics[name] = float64(value.(uint32))
	case reflect.Int64:
		if s.metrics[name] == nil {
			s.metrics[name] = value
		} else {
			s.metrics[name] = s.metrics[name].(int64) + value.(int64)
		}
	default:
		return fmt.Errorf("invalid metric type %s", reflect.TypeOf(value).Name())
	}
	log.Printf("metrics added %s,size=%d", name, len(s.metrics))
	return nil
}

func (s MemRepository) GetAllAsString() string {
	s.mx.Lock()
	defer s.mx.Unlock()
	result := ""
	for key, value := range s.metrics {
		if result != "" {
			result = result + "\n"
		}
		result = result + fmt.Sprintf("%s %+v", key, value)
	}
	return result
}

func (s MemRepository) GetAll() map[string]any {
	s.mx.Lock()
	defer s.mx.Unlock()
	var result = make(map[string]any, len(s.metrics))
	for k, v := range s.metrics {
		result[k] = v
	}
	return result
}

func (s MemRepository) Get(name string) (value any, exists bool) {
	s.mx.Lock()
	defer s.mx.Unlock()
	value, exists = s.metrics[name]
	return value, exists
}

func (s MemRepository) SaveAll(m map[string]any) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.metrics = m
	log.Printf("Metrics saved=%d", len(s.metrics))
	return nil
}
