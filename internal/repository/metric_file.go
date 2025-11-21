package repository

import (
	"encoding/json"
	"fmt"
	"io"
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
		log.Err(err).Msg("Get: GetAll return exception")
		return nil, err
	} else {
		m, ok := met[name]
		log.Debug().Msg(fmt.Sprintf("Get: search in map m=%+v,ok=%t", m, ok))
		if ok {
			return &m, err
		} else {
			return nil, nil
		}
	}
}

func (r MetricRepositoryFile) GetAll() (m map[string]model.Metric, err error) {
	f, err := os.OpenFile(r.fileName, os.O_CREATE|os.O_APPEND|os.O_RDONLY, 0666)
	if err != nil {
		log.Err(err).Msg("GetAll: file create exception")
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		log.Err(err).Msg("GetAll: file read exception")
		return nil, err
	}
	if len(b) == 0 {
		b = []byte("null")
	}

	var a []model.Metric
	log.Debug().Msg(fmt.Sprintf("GetAll: []byte from file=%s", b))
	err = json.Unmarshal(b, &a)
	if err != nil {
		log.Err(err).Msg("GetAll: json unmarshal exception")
		return nil, err
	}

	m = make(map[string]model.Metric, len(a))
	for _, item := range a {
		m[item.ID] = item
		log.Debug().Msg(fmt.Sprintf("GetAll: to map=%+v", item))
	}

	return m, nil
}

func (r MetricRepositoryFile) SetOne(m model.Metric) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	met, err := r.GetAll()
	if err != nil {
		return err
	}

	met[m.ID] = m

	var a []model.Metric
	for _, item := range met {
		a = append(a, item)
	}

	b, err := json.Marshal(&a)
	if err != nil {
		log.Err(err).Msg("SetOne: json marshal exception")
		return err
	}
	err = os.WriteFile(r.fileName, b, 0600)
	if err != nil {
		log.Err(err).Msg("SetOne: file create exception")
		return err
	}

	log.Debug().Msg(fmt.Sprintf("SetOne: metrics added %s", m.ID))
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

	var a []model.Metric
	for _, item := range met {
		a = append(a, item)
	}
	b, err := json.Marshal(&a)
	if err != nil {
		log.Err(err).Msg("SetList: json marshal exception")
		return err
	}
	err = os.WriteFile(r.fileName, b, 0600)
	if err != nil {
		log.Err(err).Msg("SetList: file create exception")
		return err
	}
	log.Debug().Msg(fmt.Sprintf("SetList: Metrics saved=%d", len(m)))
	return nil
}
