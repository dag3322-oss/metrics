package model

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/rs/zerolog/log"
)

const (
	Counter string = "counter"
	Gauge   string = "gauge"
)

const (
	ActionUpdate string = "update"
	ActionGet    string = "value"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metric struct {
	ID    string   `json:"id" db:"id"`
	MType string   `json:"type" db:"type"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
	Value *float64 `json:"value,omitempty" db:"value"`
	Hash  string   `json:"hash,omitempty" db:"hash"`
}

func FromKeyValue(m *Metric, k string, v any) error {
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

func ToKeyValue(m *Metric) (k string, v any, err error) {
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

func Validate(action string, m *Metric) (code int, err error) {
	code = http.StatusOK
	if m.ID == "" {
		code = http.StatusNotFound
		err = fmt.Errorf("empty metric name")
	}
	if code == http.StatusOK && !(m.MType == Gauge || m.MType == Counter) {
		code = http.StatusBadRequest
		err = fmt.Errorf("invalid metric type=%s", m.MType)
	}
	if code == http.StatusOK && action == ActionUpdate && (m.MType == Gauge && m.Value == nil || m.MType == Counter && m.Delta == nil) {
		code = http.StatusBadRequest
		err = fmt.Errorf("metric value not assigned name=%s, type=%s", m.ID, m.MType)
	}
	return code, err
}

func (m Metric) StringValue() string {
	switch m.MType {
	case Gauge:
		return fmt.Sprintf("%v", *m.Value)
	case Counter:
		return fmt.Sprintf("%v", *m.Delta)
	default:
		return ""
	}
}

func SetValue(m *Metric, value any) error {
	if m == nil {
		return errors.New("metric.SetValue: empty metric pointer")
	}
	switch m.MType {
	case Gauge:
		f := value.(float64)
		m.Value = &f
	case Counter:
		i := value.(int64)
		m.Delta = &i
	default:
		return fmt.Errorf("metric.SetValue: invalid metric type=%s", m.MType)
	}
	return nil
}
