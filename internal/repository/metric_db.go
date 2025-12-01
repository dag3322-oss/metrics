package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"time"

	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	pg_pool "github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type MetricRepositoryDB struct {
	context context.Context
	pool    *pg_pool.Pool
}

func NewDBRepository(connectionString *string, context context.Context) (repo *MetricRepositoryDB, err error) {
	var pool *pg_pool.Pool

	config, err := pg_pool.ParseConfig(*connectionString)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 10
	config.MaxConnLifetime = 30 * time.Second
	pool, err = pg_pool.NewWithConfig(context, config)
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

	return &MetricRepositoryDB{pool: pool, context: context}, nil
}

func (r MetricRepositoryDB) acquire() (conn *pg_pool.Conn, err error) {
	i := 0
	for {
		time.Sleep(time.Duration(i) * time.Second)
		conn, err = r.pool.Acquire(r.context)
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) {
				switch i {
				case 0:
					i = 1
				default:
					i = i + 2
				}
				if i <= 5 {
					log.Debug().Msg(fmt.Sprintf("repeat after timeout delay=%d", i))
					continue
				}
			} else {
				log.Debug().Msg(fmt.Sprintf("not network error=%+v,%s", err, reflect.TypeOf(err).Name()))
			}
			log.Err(err).Msg("request send exception")
			return nil, err
		}
		return conn, nil
	}

}

func (r MetricRepositoryDB) Get(name string) (m *model.Metric, err error) {
	conn, err := r.acquire()
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	row := conn.QueryRow(r.context, "select * from metrics_get($1)", fmt.Sprintf("{\"id\": \"%s\"}", name))
	log.Debug().Any("row", row).Msg("")
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
	conn, err := r.acquire()
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
	conn, err := r.acquire()
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
	conn, err := r.acquire()
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

func (r MetricRepositoryDB) Close() {
	r.pool.Close()
}

func (r MetricRepositoryDB) Ping() error {
	return r.pool.Ping(r.context)
}
