package transactor

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QueryEngine interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type QueryEngineProvider interface {
	GetQueryEngine(ctx context.Context) QueryEngine
}

type Transactor struct {
	pool *pgxpool.Pool
}

const key = "tx"

func NewTransactor(db *pgxpool.Pool) *Transactor {
	return &Transactor{pool: db}
}

func (t *Transactor) RunRepeatableRead(ctx context.Context, f func(ctxTX context.Context) error) error {
	tx, errTx := t.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if errTx != nil {
		return fmt.Errorf("transactor.RunRepeatableRead error: %w", errTx)
	}

	if errF := f(context.WithValue(ctx, key, tx)); errF != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return fmt.Errorf("transactor.RunRepeatableRead error: %w, rollback error: %v", errF, errRollback)
		}
		return fmt.Errorf("transactor.RunRepeatableRead error: %w", errF)
	}

	if errCommit := tx.Commit(ctx); errCommit != nil {
		return fmt.Errorf("transactor.RunRepeatableRead error: %w", errCommit)
	}

	return nil
}

func (t *Transactor) GetQueryEngine(ctx context.Context) QueryEngine {
	tx, ok := ctx.Value(key).(QueryEngine)
	if ok && tx != nil {
		return tx
	}

	return t.pool
}
