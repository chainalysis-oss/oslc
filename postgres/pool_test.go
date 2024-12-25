package postgres

import (
	"context"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewPool(t *testing.T) {
	// We don't expect the pool to verify the connection string when created, so we can just pass a bogus connection string.
	p, err := NewPool(context.Background(), "postgres://postgres:postgres@localhost:5432/postgres")
	require.NoError(t, err)
	p.Close()
}

func TestNewPoolErr(t *testing.T) {
	_, err := NewPool(context.Background(), "mysql://")
	require.Error(t, err)
}

func TestPool_Begin(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()
	mock.ExpectBegin().Times(1)
	p := &pool{
		pool: mock,
	}
	_, err = p.Begin(context.Background())
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPool_Query(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()
	mock.ExpectQuery("SELECT 1").WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1)).Times(1)
	p := &pool{
		pool: mock,
	}
	_, err = p.Query(context.Background(), "SELECT 1")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPool_Close(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()
	mock.ExpectClose().Times(1)
	p := &pool{
		pool: mock,
	}
	p.Close()
	require.NoError(t, mock.ExpectationsWereMet())
}
