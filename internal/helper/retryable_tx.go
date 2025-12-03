package helper

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func NewRetryableTx(pool *pgxpool.Pool, context context.Context) (rtx *RetryableTx, err error) {
	rtx = new(RetryableTx)
	rtx.ctx = context
	rtx.retryDelay = new(int)

	for {
		rtx.tx, err = pool.Begin(context)
		var retryErr = new(retryError)
		err = rtx.doWithRetry(err)
		if err == nil || !errors.As(err, retryErr) {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("pool.Begin: %w", err)
	} else {
		return rtx, nil
	}
}

type RetryableTx struct {
	tx         pgx.Tx
	ctx        context.Context
	retryDelay *int
}

func NewRetryError() error {
	return retryError{}
}

type retryError struct {
}

func (e retryError) Error() string {
	return "do retry"
}

func (rtx RetryableTx) doWithRetry(err error) (result error) {
	if err == nil {
		return nil
	}

	time.Sleep(time.Duration(*rtx.retryDelay) * time.Second)

	result = err
	var ne net.Error
	if errors.As(err, &ne) {
		result = fmt.Errorf("network error: %w", err)
		switch *rtx.retryDelay {
		case 0:
			*rtx.retryDelay = 1
		default:
			*rtx.retryDelay = *rtx.retryDelay + 2
		}
		if *rtx.retryDelay <= 5 {
			log.Info().Err(err).Int("delay", *rtx.retryDelay).Msg("will retry sql")
			result = NewRetryError()
		}
	} else {
		result = fmt.Errorf("not network error: %w", err)
	}
	return result
}

func (rtx RetryableTx) Query(sql string, args ...any) (rows pgx.Rows, err error) {
	ctxt, f := context.WithTimeout(rtx.ctx, 30*time.Second)
	defer f()
	defer rtx.tx.Rollback(ctxt)
	for {
		rows, err = rtx.tx.Query(ctxt, sql, args...)
		var retryErr = new(retryError)
		err = rtx.doWithRetry(err)
		if err == nil || !errors.As(err, retryErr) {
			break
		}
	}
	if err == nil {
		rtx.tx.Commit(ctxt)
		return rows, nil
	} else {
		rtx.tx.Rollback(ctxt)
		return nil, err
	}
}

func (rtx RetryableTx) Exec(sql string, args ...any) (err error) {
	ctxt, f := context.WithTimeout(rtx.ctx, 30*time.Second)
	defer f()
	defer rtx.tx.Rollback(ctxt)
	log.Debug().Str("sql", sql).Msg("")

	for {
		_, err = rtx.tx.Exec(ctxt, sql, args...)
		var retryErr = new(retryError)
		err = rtx.doWithRetry(err)
		if err == nil || !errors.As(err, retryErr) {
			break
		}
	}
	if err != nil {
		rtx.tx.Rollback(ctxt)
		return fmt.Errorf("from exec: %w", err)
	} else {
		rtx.tx.Commit(ctxt)
		return nil
	}
}

func (rtx RetryableTx) Batch(batch *pgx.Batch) (err error) {
	ctxt, f := context.WithTimeout(rtx.ctx, 30*time.Second)
	defer f()
	defer rtx.tx.Rollback(ctxt)
	br := rtx.tx.SendBatch(ctxt, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		for {
			_, err := br.Exec()
			var retryErr = new(retryError)
			err = rtx.doWithRetry(err)
			if err == nil || !errors.As(err, retryErr) {
				break
			}
		}
		if err != nil {
			rtx.tx.Rollback(ctxt)
			return fmt.Errorf("batch exec: %d %w", i, err)
		}
	}
	br.Close()
	rtx.tx.Commit(ctxt)
	return nil
}

func (rtx RetryableTx) Close() {
	ctxt, f := context.WithTimeout(rtx.ctx, 30*time.Second)
	defer f()
	rtx.tx.Conn().Close(ctxt)
}
