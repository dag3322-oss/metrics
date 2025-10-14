package metrics

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
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
	switch value.(type) {
	case float64:
		s.metrics[name] = value
	case uint64:
		s.metrics[name] = float64(value.(uint64))
	case int64:
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
	var result map[string]any = make(map[string]any, len(s.metrics))
	for k, v := range s.metrics {
		result[k] = v
	}
	return result
}
