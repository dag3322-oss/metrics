package handler

import (
	"encoding/json"
	"fmt"
	"os"

	models "github.com/dag3322-oss/metrics/internal/model"
	repository "github.com/dag3322-oss/metrics/internal/repository"
	"github.com/rs/zerolog/log"
)

func Flush(repo repository.Repository, fileName string) error {
	var err error
	var m models.Metrics
	var mkv = repo.GetAll()
	var mm []models.Metrics
	var i = 0
	for k, v := range mkv {
		err = models.FromKeyValue(&m, k, v)
		if err == nil {
			mm = append(mm, m)
		} else {
			log.Err(err).Msg("model create exception")
			return err
		}
		i++
	}
	b, err := json.Marshal(mm)
	if err != nil {
		log.Err(err).Msg("json marshal exception")
		return err
	}
	err = os.WriteFile(fileName, b, 0600)
	if err != nil {
		log.Err(err).Msg("file create exception")
		return err
	}
	log.Debug().Msg(fmt.Sprintf("flushed=%d to %s", len(mm), fileName))
	return nil
}

func Load(repo repository.Repository, fileName string) error {
	var mkv = make(map[string]any)
	var mm []models.Metrics

	b, err := os.ReadFile(fileName)
	if err != nil {
		log.Err(err).Msg("file create exception")
		return err
	}

	err = json.Unmarshal(b, &mm)
	if err != nil {
		log.Err(err).Msg("json marshal exception")
		return err
	}

	for _, m := range mm {
		k, v, err := models.ToKeyValue(&m)
		if err == nil {
			mkv[k] = v
		} else {
			log.Err(err).Msg(fmt.Sprintf("invalid metrics model=%+v", m))
			return err
		}
	}
	log.Debug().Msg(fmt.Sprintf("loaded=%d from %s,repo=%+v", len(mm), fileName, repo))
	return repo.SaveAll(mkv)
}
