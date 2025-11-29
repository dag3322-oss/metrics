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
	defer conn.Release()
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

func (r MetricRepositoryDB) GetAll() (result map[string]model.Metric, err error) {
	conn, err := r.pool.Acquire(r.context)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	row := conn.QueryRow(r.context, "select * from metrics_get(null)")
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
	result = make(map[string]model.Metric, len(mm))
	for _, m := range mm {
		result[m.ID] = m
	}
	return result, nil
}

func (r MetricRepositoryDB) SetOne(m model.Metric) error {
	conn, err := r.pool.Acquire(r.context)
	if err != nil {
		return err
	}
	defer conn.Release()
	var mm []model.Metric
	mm = append(mm, m)
	b, err := json.Marshal(&mm)
	if err != nil {
		log.Err(err).Msg("json marshal exception")
		return err
	}

	_, err = conn.Exec(r.context, "select * from metrics_set($1)", b)
	if err != nil {
		log.Err(err).Msg("sql set error")
		return err
	}
	return nil
}

func (r MetricRepositoryDB) SetList(m map[string]model.Metric) error {
	conn, err := r.pool.Acquire(r.context)
	if err != nil {
		return err
	}
	defer conn.Release()

	var a []model.Metric
	for _, item := range m {
		a = append(a, item)
	}

	b, err := json.Marshal(&a)
	if err != nil {
		log.Err(err).Msg("json marshal exception")
		return err
	}

	_, err = conn.Exec(r.context, "select * from metrics_set($1)", b)
	if err != nil {
		log.Err(err).Msg("sql set error")
		return err
	}
	return nil
}
