package handler

import (
	"github.com/dag3322-oss/metrics/internal/repository"
)

func Flush(repoMem repository.Metric, repoFile repository.Metric) error {
	m, err := repoMem.GetAll()
	if err == nil {
		err = repoFile.SetList(m)
	}
	return err
}

func Load(repoMem repository.Metric, repoFile repository.Metric) error {
	m, err := repoFile.GetAll()
	if err == nil {
		err = repoMem.SetList(m)
	}
	return err
}
