package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dag3322-oss/metrics/internal/helper"
	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	pg_pool "github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const sqlSet = `
	insert into metric (id, "type", delta, "value", hash) values ($1, $2, $3, $4, $5) on conflict (id) do update
	set delta = case when excluded."type" = 'counter' then coalesce(metric.delta, 0) + excluded.delta else excluded.delta end, 
	"value" = excluded."value", 
	hash = excluded.hash
	`

type MetricRepositoryDB struct {
	context context.Context
	pool    *pg_pool.Pool
}

func NewDBRepository(connectionString *string, ctx context.Context) (repo *MetricRepositoryDB, err error) {
	var pool *pg_pool.Pool

	config, err := pg_pool.ParseConfig(*connectionString)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 10
	config.MaxConnLifetime = 20 * time.Second
	ctxt, f := context.WithTimeout(ctx, 30*time.Second)
	defer f()
	pool, err = pg_pool.NewWithConfig(ctxt, config)
	if err != nil {
		return nil, err
	}

	log.Debug().Msg("Database migrations will be applied")
	driver, err := iofs.New(migrations.FS, "sql")
	if err != nil {
		return nil, err
	}
	log.Debug().Msg("iofs driver created")

	m, err := migrate.NewWithSourceInstance("iofs", driver, *connectionString)
	if err != nil {
		return nil, err
	}
	defer m.Close()
	log.Debug().Msg("migration isnstance created")

	log.Err(err)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, err
	}
	log.Debug().Msg("Database migrations applied succesfully")

	return &MetricRepositoryDB{pool: pool, context: ctx}, nil
}

func (r MetricRepositoryDB) Get(name string) (m *model.Metric, err error) {
	tx, err := helper.NewRetryableTx(r.pool, r.context)
	if err != nil {
		return nil, fmt.Errorf("from NewRetryableTx: %w", err)
	}
	rows, err := tx.Query("select id, \"type\", delta, \"value\", coalesce(hash, '') hash from metric where id = $1", name)
	if err != nil {
		return nil, fmt.Errorf("from query: %w", err)
	}
	defer rows.Close()
	log.Debug().Msg("before next")
	if rows.Next() {
		log.Debug().Any("row", rows).Msg("")
		m, err := pgx.RowToStructByNameLax[model.Metric](rows)
		if err != nil {
			return nil, fmt.Errorf("from rowToStructByNameLax: %w", err)
		}
		return &m, nil
	} else {
		return nil, nil
	}
}

func (r MetricRepositoryDB) GetAll() (result map[string]model.Metric, err error) {
	tx, err := helper.NewRetryableTx(r.pool, r.context)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query("select id, \"type\", delta, \"value\", coalesce(hash, '') hash from metric")
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()
	a, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Metric])
	if err != nil {
		return nil, fmt.Errorf("CollectRows: %w", err)
	}
	result = make(map[string]model.Metric)
	for _, m := range a {
		result[m.ID] = m
	}
	return result, nil
}

func (r MetricRepositoryDB) SetOne(m model.Metric) error {
	tx, err := helper.NewRetryableTx(r.pool, r.context)
	if err != nil {
		return err
	}
	err = tx.Exec(sqlSet, m.ID, m.MType, m.Delta, m.Value, m.Hash)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (r MetricRepositoryDB) SetList(m map[string]model.Metric) error {
	batch := &pgx.Batch{}
	for _, item := range m {
		batch.Queue(sqlSet, item.ID, item.MType, item.Delta, item.Value, item.Hash)
	}

	tx, err := helper.NewRetryableTx(r.pool, r.context)
	if err != nil {
		return err
	}
	err = tx.Batch(batch)
	if err != nil {
		return fmt.Errorf("batch: %w", err)
	}
	return nil
}

func (r MetricRepositoryDB) Close() {
	r.pool.Close()
}

func (r MetricRepositoryDB) Ping() error {
	return r.pool.Ping(r.context)
}
