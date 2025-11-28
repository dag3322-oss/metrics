package repository

import (
	"fmt"
	"sync"

	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/rs/zerolog/log"
)

type MetricRepositoryMem struct {
	mx      *sync.Mutex
	metrics map[string]*model.Metric
}

func NewMemRepository() MetricRepositoryMem {
	return MetricRepositoryMem{mx: &sync.Mutex{}, metrics: make(map[string]*model.Metric)}
}

func (r MetricRepositoryMem) Get(name string) (model *model.Metric, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	m, ok := r.metrics[name]
	if ok {
		return m, nil
	} else {
		return nil, nil
	}
}

func (r MetricRepositoryMem) GetAll() (m map[string]model.Metric, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	var result = make(map[string]model.Metric, len(r.metrics))
	for _, m := range r.metrics {
		result[m.ID] = *m
	}
	return result, nil
}

func (r MetricRepositoryMem) SetOne(m model.Metric) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if metricSaved, ok := r.metrics[m.ID]; ok && m.MType == model.Counter && metricSaved.Delta != nil {
		*m.Delta = *m.Delta + *metricSaved.Delta
	}
	r.metrics[m.ID] = &m

	log.Debug().Msg(fmt.Sprintf("metrics added %s,%s,%+v,%+v", r.metrics[m.ID].ID, r.metrics[m.ID].MType, r.metrics[m.ID].Value, r.metrics[m.ID].Delta))

	return nil
}

func (r MetricRepositoryMem) SetList(m map[string]model.Metric) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	for _, metricToSave := range m {
		if metricSaved, ok := r.metrics[metricToSave.ID]; ok && metricToSave.MType == model.Counter && metricSaved.Delta != nil {
			*metricToSave.Delta = *metricToSave.Delta + *metricSaved.Delta
		}
		r.metrics[metricToSave.ID] = &metricToSave
	}

	log.Debug().Msg(fmt.Sprintf("Metrics saved=%d", len(m)))

	return nil
}
