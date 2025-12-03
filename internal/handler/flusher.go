package handler

import (
	"github.com/dag3322-oss/metrics/internal/repository"
)

func Flush(repo repository.Metric, repoFile repository.Metric) error {
	m, err := repo.GetAll()
	if err == nil {
		err = repoFile.SetList(m)
	}
	return err
}

func Load(repo repository.Metric, repoFile repository.Metric) error {
	m, err := repoFile.GetAll()
	if err == nil {
		err = repo.SetList(m)
	}
	return err
}
