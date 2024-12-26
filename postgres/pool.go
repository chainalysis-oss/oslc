package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is an interface for a database pool. It's sole purpose is to abstract away the underlying implementation and
// allow for easy mocking, thus facilitating unit testutils. The [NewPool] function is used to create a new pool for
// code intended to be used in production.
type Pool interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

// pool is an implementation of the [Pool] interface. It is a private type and should not be referenced directly.
// Instead, refer to it through the [Pool] interface.
type pool struct {
	pool Pool
}

// NewPool creates a new database connection pool for use in production code. It returns a private type, that
// implements the [Pool] interface.
func NewPool(ctx context.Context, connString string) (Pool, error) {
	var p pool
	var err error
	p.pool, err = pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Query executes a query against the database. It wraps the underlying [pgxpool.Pool.Query] function.
func (p *pool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

// Begin starts a transaction. It wraps the underlying [pgxpool.Pool.Begin] function.
func (p *pool) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.pool.Begin(ctx)
}

// Close closes the pool. It wraps the underlying [pgxpool.Pool.Close] function.
func (p *pool) Close() {
	p.pool.Close()
}
