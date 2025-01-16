package oslc

import (
	"context"
	"errors"
	"github.com/chainalysis-oss/oslc"
	oslcv1alpha "github.com/chainalysis-oss/oslc/gen/oslc/oslc/v1alpha"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

type Server struct {
	options *serverOptions
	oslcv1alpha.UnimplementedOslcServiceServer
}

type InvalidDistributorError struct {
	Distributor string
}

func (e InvalidDistributorError) Error() string {
	return "invalid distributor: " + e.Distributor
}

func (s Server) getPackageFromDistributor(ctx context.Context, distributor string, name string, version string) (oslc.Entry, error) {
	var entry oslc.Entry
	var err error
	switch distributor {
	case oslc.DistributorPypi:
		entry, err = s.options.PypiClient.GetPackageVersion(name, version)
	case oslc.DistributorNpm:
		entry, err = s.options.NpmClient.GetPackageVersion(name, version)
	case oslc.DistributorMaven:
		entry, err = s.options.MavenClient.GetPackageVersion(name, version)
	case oslc.DistributorCratesIo:
		entry, err = s.options.CratesIoClient.GetPackageVersion(name, version)
	case oslc.DistributorGo:
		entry, err = s.options.GoClient.GetPackageVersion(name, version)
	default:
		return oslc.Entry{}, InvalidDistributorError{Distributor: distributor}
	}
	if err != nil {
		return oslc.Entry{}, err
	}

	entry = s.normalizeEntry(ctx, entry)

	return entry, nil
}

func (s Server) normalizeEntry(ctx context.Context, entry oslc.Entry) oslc.Entry {
	lic := s.options.LicenseIDNormalizer.NormalizeID(ctx, entry.License)
	entry.License = lic
	return entry
}

func (s Server) GetPackageInfo(ctx context.Context, request *oslcv1alpha.GetPackageInfoRequest) (*oslcv1alpha.GetPackageInfoResponse, error) {
	if !validDistributor(request.Distributor) {
		return nil, status.Error(codes.InvalidArgument, "invalid distributor")
	}

	var entry oslc.Entry
	var err error
	entry, err = s.options.Datastore.Retrieve(ctx, request.Name, request.Version, request.Distributor)
	if err != nil {
		if errors.Is(err, oslc.ErrDatastoreObjectNotFound) {
			s.options.Logger.DebugContext(ctx, "package not found in datastore, querying upstream")
		} else {
			s.options.Logger.Error("failed to retrieve from datastore", slog.String("error", err.Error()))
		}

		entry, err = s.getPackageFromDistributor(ctx, request.Distributor, request.Name, request.Version)

		if err != nil {
			s.options.Logger.Error("failed to retrieve from upstream", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "internal server error")
		}

		if err := s.options.Datastore.Save(ctx, entry); err != nil {
			s.options.Logger.Error("failed to save to datastore", slog.String("error", err.Error()))
		}
	}
	dps := make([]*oslcv1alpha.DistributionPoint, len(entry.DistributionPoints))
	for i, dp := range entry.DistributionPoints {
		dps[i] = &oslcv1alpha.DistributionPoint{
			Name:        dp.Name,
			Url:         dp.URL,
			Distributor: dp.Distributor,
		}
	}
	return &oslcv1alpha.GetPackageInfoResponse{
		Name:               entry.Name,
		Version:            entry.Version,
		License:            entry.License,
		DistributionPoints: dps,
	}, nil
}

func NewServer(options ...ServerOption) (*Server, error) {
	opts := defaultServerOptions
	for _, opt := range globalServerOptions {
		opt.apply(&opts)
	}
	for _, opt := range options {
		opt.apply(&opts)
	}

	return &Server{
		options: &opts,
	}, nil
}

func validDistributor(distributor string) bool {
	return distributor == oslc.DistributorPypi || distributor == oslc.DistributorNpm || distributor == oslc.DistributorMaven || distributor == oslc.DistributorCratesIo || distributor == oslc.DistributorGo
}
