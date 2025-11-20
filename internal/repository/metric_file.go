package repository

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"sync"

	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/rs/zerolog/log"
)

type MetricRepositoryFile struct {
	mx       *sync.Mutex
	fileName string
}

func NewFileRepository(fileName string) MetricRepositoryFile {
	return MetricRepositoryFile{mx: &sync.Mutex{}, fileName: fileName}
}

func (r MetricRepositoryFile) Get(name string) (*model.Metric, error) {
	met, err := r.GetAll()
	if err != nil {
		return nil, err
	} else {
		m, ok := met[name]
		if ok {
			return &m, err
		} else {
			return nil, nil
		}
	}
}

func (r MetricRepositoryFile) GetAll() (m map[string]model.Metric, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	var result map[string]model.Metric
	var mm []model.Metric

	b, err := os.ReadFile(r.fileName)
	if err != nil {
		log.Err(err).Msg("file create exception")
		return nil, err
	}

	err = json.Unmarshal(b, &mm)
	if err != nil {
		log.Err(err).Msg("json marshal exception")
		return nil, err
	}

	result = make(map[string]model.Metric, len(mm))
	for _, met := range mm {
		result[met.ID] = met
	}

	return result, nil
}

func (r MetricRepositoryFile) SetOne(m model.Metric) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	met, err := r.GetAll()
	if err != nil {
		return err
	}

	met[m.ID] = m

	b, err := json.Marshal(met)
	if err != nil {
		log.Err(err).Msg("json marshal exception")
		return err
	}
	err = os.WriteFile(r.fileName, b, 0600)
	if err != nil {
		log.Err(err).Msg("file create exception")
		return err
	}

	log.Debug().Msg(fmt.Sprintf("metrics added %s", m.ID))
	return nil
}

func (r MetricRepositoryFile) SetList(m map[string]model.Metric) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	met, err := r.GetAll()
	if err != nil {
		return err
	}
	maps.Copy(met, m)
	b, err := json.Marshal(met)
	if err != nil {
		log.Err(err).Msg("json marshal exception")
		return err
	}
	err = os.WriteFile(r.fileName, b, 0600)
	if err != nil {
		log.Err(err).Msg("file create exception")
		return err
	}
	log.Debug().Msg(fmt.Sprintf("Metrics saved=%d", len(m)))
	return nil
}
