package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dag3322-oss/metrics/internal/model"
	pg_pool "github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type MetricRepositoryDB struct {
	context context.Context
	pool    *pg_pool.Pool
}

func NewDBRepository(pool *pg_pool.Pool, context context.Context) MetricRepositoryDB {
	return MetricRepositoryDB{pool: pool, context: context}
}

func (r MetricRepositoryDB) Get(name string) (m *model.Metric, err error) {
	conn, err := r.pool.Acquire(r.context)
	if err != nil {
		return nil, err
	}
	row := conn.QueryRow(r.context, "select * from metrics_get($1)", fmt.Sprintf("{\"id\": \"%s\"}", name))
	log.Debug().Msg(fmt.Sprintf("db row=%+v", row))
	var b []byte
	err = row.Scan(&b)
	if err != nil {
		log.Err(err).Msg("invalid row type")
		return nil, err
	}

	var mm []model.Metric
	err = json.Unmarshal(b, &mm)
	if err != nil {
		log.Err(err).Msg("json marshal exception")
		return nil, err
	}
	if len(mm) > 0 {
		return &mm[0], nil
	} else {
		return nil, nil
	}
}

func (s MetricRepositoryDB) GetAll() (m map[string]model.Metric, err error) {
	return nil, nil
}

func (s MetricRepositoryDB) SetOne(m model.Metric) error {
	return nil
}

func (s MetricRepositoryDB) SetList(m map[string]model.Metric) error {
	return nil
}
