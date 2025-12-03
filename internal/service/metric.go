package service

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/dag3322-oss/metrics/internal/model"
)

func NameValueToModel(name string, value any) (m *model.Metric, err error) {
	if name == "" {
		return nil, errors.New("empty metric name")
	}
	if value == nil {
		return nil, errors.New("empty metric value")
	}
	m = new(model.Metric)
	m.ID = name
	switch reflect.ValueOf(value).Kind() {
	case reflect.Float64:
		f := float64(value.(float64))
		m.MType = model.Gauge
		m.Value = &f
	case reflect.Uint64:
		f := float64(value.(uint64))
		m.MType = model.Gauge
		m.Value = &f
	case reflect.Uint32:
		f := float64(value.(uint32))
		m.MType = model.Gauge
		m.Value = &f
	case reflect.Int64:
		i := int64(value.(int64))
		m.MType = model.Counter
		m.Delta = &i
	default:
		return nil, fmt.Errorf("invalid metric type %s", reflect.TypeOf(value).Name())
	}
	return m, nil
}
