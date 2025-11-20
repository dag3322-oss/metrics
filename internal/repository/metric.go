package repository

import (
	"github.com/dag3322-oss/metrics/internal/model"
)

type Metric interface {
	Get(name string) (m *model.Metric, err error)
	GetAll() (m map[string]model.Metric, err error)
	SetOne(m model.Metric) error
	SetList(m map[string]model.Metric) error
}
