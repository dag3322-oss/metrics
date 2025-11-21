package repository

import (
	"fmt"
	"maps"
	"sync"

	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/internal/service"
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

func (r MetricRepositoryMem) GetAll() (m map[string]model.Metric, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	var result = make(map[string]model.Metric, len(r.metrics))
	maps.Copy(result, r.metrics)
	return result, nil
}

func (r MetricRepositoryMem) SetOne(m model.Metric) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	item, ok := r.metrics[m.ID]
	if ok {
		err := service.UpdateMetricValue(&item, &m)
		if err != nil {
			log.Err(err).Msg("SetOne: UpdateMetricValue")
			return err
		}
	} else {
		r.metrics[m.ID] = m
	}

	log.Debug().Msg(fmt.Sprintf("metrics added %s", m.ID))
	return nil
}

func (r MetricRepositoryMem) SetList(m map[string]model.Metric) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	for _, itemSrc := range m {
		itemDst, ok := r.metrics[itemSrc.ID]
		if ok {
			err := service.UpdateMetricValue(&itemDst, &itemSrc)
			if err != nil {
				log.Err(err).Msg("SetOne: UpdateMetricValue")
				return err
			}
		} else {
			r.metrics[itemSrc.ID] = itemSrc
		}
	}

	log.Debug().Msg(fmt.Sprintf("Metrics saved=%d", len(m)))
	return nil
}
