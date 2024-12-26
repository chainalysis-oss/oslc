package postgres

import (
	"context"
	"errors"
	"github.com/chainalysis-oss/oslc"
	"github.com/jackc/pgx/v5"
	"log/slog"
)

type Datastore struct {
	options *datastoreOptions
}

func NewDatastore(options ...DatastoreOption) (*Datastore, error) {
	opts := defaultDatastoreOptions
	for _, opt := range globalDatastoreOptions {
		opt.apply(&opts)
	}
	for _, opt := range options {
		opt.apply(&opts)
	}

	if opts.Pool == nil {
		return nil, ErrMissingOptionPool
	}

	return &Datastore{
		options: &opts,
	}, nil
}

type datastoreOptions struct {
	Logger *slog.Logger
	Pool   Pool
}

var defaultDatastoreOptions = datastoreOptions{
	Logger: slog.Default(),
}

var globalDatastoreOptions []DatastoreOption

type DatastoreOption interface {
	apply(*datastoreOptions)
}

type funcDatastoreOption struct {
	f func(*datastoreOptions)
}

func (fdo *funcDatastoreOption) apply(opts *datastoreOptions) {
	fdo.f(opts)
}

func newFuncDatastoreOption(f func(*datastoreOptions)) *funcDatastoreOption {
	return &funcDatastoreOption{
		f: f,
	}
}

func WithLogger(logger *slog.Logger) DatastoreOption {
	return newFuncDatastoreOption(func(opts *datastoreOptions) {
		opts.Logger = logger
	})
}

func WithPool(pool Pool) DatastoreOption {
	return newFuncDatastoreOption(func(opts *datastoreOptions) {
		opts.Pool = pool
	})
}

var datastoreSaveStatement = "INSERT INTO packages (name, license, version, distributor, distribution_url) VALUES ($1, $2, $3, $4, $5) ON CONFLICT ON CONSTRAINT packages_pk DO UPDATE SET license = $2, distribution_url = $5"

func (d *Datastore) Save(ctx context.Context, entry oslc.Entry) error {
	tx, err := d.options.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, dp := range entry.DistributionPoints {
		_, err = tx.Exec(ctx, datastoreSaveStatement, entry.Name, entry.License, entry.Version, dp.Distributor, dp.URL)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

var datastoreRetrieveStatement = "SELECT license, distribution_url FROM packages WHERE name = $1 AND version = $2 AND distributor = $3"

func (d *Datastore) Retrieve(ctx context.Context, name, version, distributor string) (oslc.Entry, error) {
	rows, err := d.options.Pool.Query(ctx, datastoreRetrieveStatement, name, version, distributor)
	if err != nil {
		return oslc.Entry{}, err
	}
	var entry oslc.Entry
	var license string
	var url string

	dp := make([]oslc.DistributionPoint, 0)
	_, err = pgx.ForEachRow(rows, []any{&license, &url}, func() error {
		dp = append(dp, oslc.DistributionPoint{
			Name:        name,
			URL:         url,
			Distributor: distributor,
		})
		return nil
	})
	if err != nil {
		return oslc.Entry{}, err
	}

	if len(dp) == 0 {
		return oslc.Entry{}, oslc.ErrDatastoreObjectNotFound
	}

	entry = oslc.Entry{
		Name:               name,
		DistributionPoints: dp,
		License:            license,
		Version:            version,
	}

	return entry, nil
}

var ErrMissingOptionPool = errors.New("missing option: pool")
