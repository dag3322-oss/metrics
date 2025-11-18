package models

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/rs/zerolog/log"
)

const (
	Counter      = "counter"
	Gauge        = "gauge"
	ActionUpdate = "update"
	ActionGet    = "value"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func FromKeyValue(m *Metrics, k string, v any) error {
	var err error

	m.ID = k
	switch vt := v.(type) {
	case float64:
		m.MType = Gauge
		if f, ok := v.(float64); ok {
			m.Value = &f
		} else {
			err = fmt.Errorf("any to float conversion error")
			log.Err(err).Msg("")
			return err
		}
	case int64:
		m.MType = Counter
		if i, ok := v.(int64); ok {
			m.Delta = &i
		} else {
			err = fmt.Errorf("any to int conversion error")
			log.Err(err).Msg("")
			return err
		}
	default:
		err = fmt.Errorf("invalid metric type %s", reflect.TypeOf(vt).Name())
		log.Err(err).Msg("")
		return err
	}
	return nil
}

func ToKeyValue(m *Metrics) (k string, v any, err error) {
	k = m.ID
	if k == "" {
		err = fmt.Errorf("empty metric name")
	} else {
		switch m.MType {
		case Gauge:
			v = *m.Value
		case Counter:
			v = *m.Delta
		default:
			err = fmt.Errorf("invalid metric type=%s", m.MType)
		}
	}
	return k, v, err
}

func Validate(action string, m *Metrics) (code int, err error) {
	code = http.StatusOK
	if m.ID == "" {
		code = http.StatusNotFound
		err = fmt.Errorf("empty metric name")
	}
	if code == http.StatusOK && !(m.MType == Gauge || m.MType == Counter) {
		code = http.StatusBadRequest
		err = fmt.Errorf("invalid metric type=%s", m.MType)
	}
	log.Printf("model validate=%+v,code=%d,err=%+v", m, code, err)
	return code, err
}

func SetValue(m *Metrics, value any) error {
	switch m.MType {
	case Gauge:
		f := value.(float64)
		m.Value = &f
	case Counter:
		i := value.(int64)
		m.Delta = &i
	default:
		return fmt.Errorf("invalid metric type=%s", m.MType)
	}
	return nil
}
