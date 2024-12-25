package postgres

import (
	"context"
	"github.com/chainalysis-oss/oslc"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"testing"
)

func newPoolMock(t *testing.T) pgxmock.PgxPoolIface {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	return mock
}

func TestNewDatastore(t *testing.T) {
	mock := newPoolMock(t)
	ds, err := NewDatastore(WithPool(mock))
	require.NoError(t, err)
	require.NotNil(t, ds)
	require.NoError(t, mock.ExpectationsWereMet())
	require.Equal(t, mock, ds.options.Pool)
}

func TestNewDatastore_ErrMissingOptionPool(t *testing.T) {
	_, err := NewDatastore()
	require.Error(t, err)
	require.Equal(t, ErrMissingOptionPool, err)
}

func TestNewDatastore_globalOptionsAreApplied(t *testing.T) {
	optCopy := make([]DatastoreOption, len(globalDatastoreOptions))
	copy(optCopy, globalDatastoreOptions)
	defer func() {
		globalDatastoreOptions = optCopy
	}()

	mock := newPoolMock(t)
	globalDatastoreOptions = append(globalDatastoreOptions, WithPool(mock))
	ds, err := NewDatastore()
	require.NoError(t, err)
	require.Equal(t, mock, ds.options.Pool)
}

func TestFuncDatastoreOption_apply(t *testing.T) {
	mock := newPoolMock(t)
	opts := datastoreOptions{}
	fdo := newFuncDatastoreOption(func(o *datastoreOptions) {
		o.Pool = mock
	})
	fdo.apply(&opts)
	require.Equal(t, mock, opts.Pool)
}

func TestNewFuncDatastoreOption(t *testing.T) {
	mock := newPoolMock(t)
	fdo := newFuncDatastoreOption(func(o *datastoreOptions) {
		o.Pool = mock
	})
	require.NotNil(t, fdo)
}

func TestWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	require.NotNil(t, logger)
	opts := datastoreOptions{}
	f := WithLogger(logger)
	f.apply(&opts)
	require.Equal(t, logger, opts.Logger)
}

func TestWithPool(t *testing.T) {
	mock := newPoolMock(t)
	opts := datastoreOptions{}
	f := WithPool(mock)
	f.apply(&opts)
	require.Equal(t, mock, opts.Pool)
}

func TestDatastore_Save(t *testing.T) {
	mock := newPoolMock(t)
	ds, err := NewDatastore(WithPool(mock))
	require.NoError(t, err)
	require.NotNil(t, ds)
	mock.ExpectBegin().Times(1)
	mock.ExpectExec(datastoreSaveStatement).
		WithArgs("test", "test4", "test5", "test3", "https://example.com").
		WillReturnResult(pgxmock.NewResult("INSERT", 1)).
		Times(1)
	mock.ExpectCommit().Times(1)
	err = ds.Save(context.Background(), oslc.Entry{
		Name: "test",
		DistributionPoints: []oslc.DistributionPoint{{
			Name:        "test2",
			URL:         "https://example.com",
			Distributor: "test3",
		}},
		License: "test4",
		Version: "test5",
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDatastore_Save_ErrBegin(t *testing.T) {
	mock := newPoolMock(t)
	ds, err := NewDatastore(WithPool(mock))
	require.NoError(t, err)
	require.NotNil(t, ds)
	mock.ExpectBegin().WillReturnError(assert.AnError)
	err = ds.Save(context.Background(), oslc.Entry{})
	require.Error(t, err)
}

func TestDatastore_Save_ErrExec(t *testing.T) {
	mock := newPoolMock(t)
	ds, err := NewDatastore(WithPool(mock))
	require.NoError(t, err)
	require.NotNil(t, ds)
	mock.ExpectBegin().Times(1)
	mock.ExpectExec(datastoreSaveStatement).
		WithArgs("test", "test4", "test5", "test3", "https://example.com").
		WillReturnError(assert.AnError)
	mock.ExpectRollback().Times(1)
	err = ds.Save(context.Background(), oslc.Entry{
		Name: "test",
		DistributionPoints: []oslc.DistributionPoint{{
			Name:        "test2",
			URL:         "https://example.com",
			Distributor: "test3",
		}},
		License: "test4",
		Version: "test5",
	})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDatastore_Save_ErrCommit(t *testing.T) {
	mock := newPoolMock(t)
	ds, err := NewDatastore(WithPool(mock))
	require.NoError(t, err)
	require.NotNil(t, ds)
	mock.ExpectBegin().Times(1)
	mock.ExpectExec(datastoreSaveStatement).
		WithArgs("test", "test4", "test5", "test3", "https://example.com").
		WillReturnResult(pgxmock.NewResult("INSERT", 1)).
		Times(1)
	mock.ExpectCommit().WillReturnError(assert.AnError)
	err = ds.Save(context.Background(), oslc.Entry{
		Name: "test",
		DistributionPoints: []oslc.DistributionPoint{{
			Name:        "test2",
			URL:         "https://example.com",
			Distributor: "test3",
		}},
		License: "test4",
		Version: "test5",
	})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDatastore_Retrieve(t *testing.T) {
	mock := newPoolMock(t)
	ds, err := NewDatastore(WithPool(mock))
	require.NoError(t, err)
	require.NotNil(t, ds)
	mock.ExpectQuery(datastoreRetrieveStatement).
		WithArgs("test", "test2", "test3").
		WillReturnRows(mock.NewRows([]string{"license", "distribution_url"}).AddRow("test4", "https://example.com")).
		Times(1)
	entry, err := ds.Retrieve(context.Background(), "test", "test2", "test3")
	require.NoError(t, err)
	require.Equal(t, oslc.Entry{
		Name: "test",
		DistributionPoints: []oslc.DistributionPoint{{
			Name:        "test",
			URL:         "https://example.com",
			Distributor: "test3",
		}},
		License: "test4",
		Version: "test2",
	}, entry)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDatastore_Retrieve_ErrQuery(t *testing.T) {
	mock := newPoolMock(t)
	ds, err := NewDatastore(WithPool(mock))
	require.NoError(t, err)
	require.NotNil(t, ds)
	mock.ExpectQuery(datastoreRetrieveStatement).
		WithArgs("test", "test2", "test3").
		WillReturnError(assert.AnError)
	_, err = ds.Retrieve(context.Background(), "test", "test2", "test3")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDatastore_Retrieve_ErrRows(t *testing.T) {
	mock := newPoolMock(t)
	ds, err := NewDatastore(WithPool(mock))
	require.NoError(t, err)
	require.NotNil(t, ds)
	mock.ExpectQuery(datastoreRetrieveStatement).
		WithArgs("test", "test2", "test3").
		// intentionally return a row that cannot be scanned, forcing the code to return an error.
		WillReturnRows(mock.NewRows([]string{"license"}).AddRow("test4")).
		Times(1)
	_, err = ds.Retrieve(context.Background(), "test", "test2", "test3")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDatastore_Retrieve_ErrNotFound(t *testing.T) {
	mock := newPoolMock(t)
	ds, err := NewDatastore(WithPool(mock))
	require.NoError(t, err)
	require.NotNil(t, ds)
	mock.ExpectQuery(datastoreRetrieveStatement).
		WithArgs("test", "test2", "test3").
		WillReturnRows(mock.NewRows([]string{"license", "distribution_url"})).
		Times(1)
	_, err = ds.Retrieve(context.Background(), "test", "test2", "test3")
	require.Error(t, err)
	require.Equal(t, oslc.ErrDatastoreObjectNotFound, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
