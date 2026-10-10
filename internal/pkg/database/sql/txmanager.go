package sql

import (
	"context"
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

// TxManager is the persistence port for use cases that must run atomically.
// It hides the concrete ORM/driver behind two operations: executing a
// function inside a transaction and reading the PostgreSQL clock.
type TxManager interface {
	// WithinTx runs fn inside a database transaction. The ctx handed to fn
	// carries the transaction, so repositories join it transparently via
	// TxOrDB. A nested WithinTx call reuses the outer transaction.
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
	// Now returns the PostgreSQL server time, honoring a transaction
	// carried by ctx.
	Now(ctx context.Context) (time.Time, error)
}

type ctxKeyTx struct{}

// BunTxManager is the Bun/pgdriver implementation of TxManager.
type BunTxManager struct {
	db *bun.DB
}

func NewBunTxManager(db *bun.DB) *BunTxManager {
	return &BunTxManager{db: db}
}

func (m *BunTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := txFromContext(ctx); ok {
		// Already inside a transaction: reuse it instead of nesting.
		return fn(ctx)
	}

	return m.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		return fn(context.WithValue(ctx, ctxKeyTx{}, tx))
	})
}

func (m *BunTxManager) Now(ctx context.Context) (time.Time, error) {
	var now time.Time
	err := TxOrDB(ctx, m.db).NewSelect().ColumnExpr("NOW()").Scan(ctx, &now)
	return now, err
}

// TxOrDB returns the transaction carried by ctx, or the fallback executor
// (usually the repository's connection pool) when ctx has none. Repositories
// call this so their queries automatically join an ambient transaction.
func TxOrDB(ctx context.Context, fallback bun.IDB) bun.IDB {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return fallback
}

func txFromContext(ctx context.Context) (bun.Tx, bool) {
	tx, ok := ctx.Value(ctxKeyTx{}).(bun.Tx)
	return tx, ok
}
