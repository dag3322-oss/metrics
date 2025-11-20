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
	switch reflect.ValueOf(value).Kind() {
	case reflect.Float64:
		*m.Value = float64(value.(float64))
	case reflect.Uint64:
		*m.Value = float64(value.(uint64))
	case reflect.Uint32:
		*m.Value = float64(value.(uint32))
	case reflect.Int64:
		*m.Delta = int64(value.(int64))
	default:
		return nil, fmt.Errorf("invalid metric type %s", reflect.TypeOf(value).Name())
	}
	return m, nil
}
