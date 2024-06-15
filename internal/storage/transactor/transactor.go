package transactor

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Transactor struct {
	Db *pgxpool.Pool
}

const key = "tx"

func (t *Transactor) RunRepeatableRead(ctx context.Context, f func(ctxTX context.Context) error) error {
	tx, errTx := t.Db.BeginTx(ctx, pgx.TxOptions{
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
