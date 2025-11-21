package repository

import (
	"fmt"
	"maps"
	"sync"

	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/rs/zerolog/log"
)

type MetricRepositoryMem struct {
	mx      *sync.Mutex
	metrics map[string]model.Metric
}

func NewMemRepository() MetricRepositoryMem {
	return MetricRepositoryMem{mx: &sync.Mutex{}, metrics: make(map[string]model.Metric)}
}

func (r MetricRepositoryMem) Get(name string) (model *model.Metric, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	m, ok := r.metrics[name]
	if ok {
		return &m, nil
	} else {
		return nil, nil
	}
}

func (s MetricRepositoryMem) GetAll() (m map[string]model.Metric, err error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	var result = make(map[string]model.Metric, len(s.metrics))
	maps.Copy(result, s.metrics)
	return result, nil
}

func (s MetricRepositoryMem) SetOne(m model.Metric) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.metrics[m.ID] = m

	log.Debug().Msg(fmt.Sprintf("metrics added %s", m.ID))
	return nil
}

func (s MetricRepositoryMem) SetList(m map[string]model.Metric) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	maps.Copy(s.metrics, m)
	log.Debug().Msg(fmt.Sprintf("Metrics saved=%d", len(m)))
	return nil
}
