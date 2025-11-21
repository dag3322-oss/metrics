package repository

import (
	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/internal/service"
	"github.com/rs/zerolog/log"
)

type MetricRepositoryMock struct {
}

func NewMockRepository() MetricRepositoryMock {
	return MetricRepositoryMock{}
}

func (r MetricRepositoryMock) Get(name string) (model *model.Metric, err error) {
	switch name {
	case "int64":
		mm, err := service.NameValueToModel("int64", int64(100500))
		if err != nil {
			log.Err(err).Msg("model from value not created,int64")
			return nil, err
		}
		return mm, err
	case "float64":
		mm, err := service.NameValueToModel("float64", float64(100500.05))
		if err != nil {
			log.Err(err).Msg("model from value not created,int64")
			return nil, err
		}
		return mm, err
	default:
		return nil, err
	}
}

func (r MetricRepositoryMock) GetAll() (m map[string]model.Metric, err error) {
	m = make(map[string]model.Metric)
	mm, _ := service.NameValueToModel("int64", 100500)
	m["int64"] = *mm
	mm, _ = service.NameValueToModel("float64", 100500.05)
	m["float64"] = *mm
	return m, nil
}

func (r MetricRepositoryMock) SetOne(m model.Metric) error {
	return nil
}

func (r MetricRepositoryMock) SetList(m map[string]model.Metric) error {
	return nil
}
