package oslc

import (
	"context"
	"github.com/chainalysis-oss/oslc"
	oslcv1alpha "github.com/chainalysis-oss/oslc/gen/oslc/oslc/v1alpha"
	oslcMocks "github.com/chainalysis-oss/oslc/mocks/oslc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
)

func Test_validDistributor(t *testing.T) {
	type args struct {
		distributor string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "pypi",
			args: args{
				distributor: oslc.DistributorPypi,
			},
			want: true,
		},
		{
			name: "npm",
			args: args{
				distributor: oslc.DistributorNpm,
			},
			want: true,
		},
		{
			name: "maven",
			args: args{
				distributor: oslc.DistributorMaven,
			},
			want: true,
		},
		{
			name: "cratesio",
			args: args{
				distributor: oslc.DistributorCratesIo,
			},
			want: true,
		},
		{
			name: "invalid",
			args: args{
				distributor: "invalid",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, validDistributor(tt.args.distributor))
		})
	}
}

var pypiRequestsEntry = oslc.Entry{
	Name:    "requests",
	Version: "2.32.3",
	License: "Apache-2.0",
	DistributionPoints: []oslc.DistributionPoint{{
		Name:        "requests",
		URL:         "https://pypi.org/project/requests/",
		Distributor: oslc.DistributorPypi,
	}},
}

var pypiRequestsGetPackageInfoRequest = oslcv1alpha.GetPackageInfoRequest{
	Name:        "requests",
	Version:     "2.32.3",
	Distributor: oslc.DistributorPypi,
}

var pypiRequestsGetPackageInfoResponse = oslcv1alpha.GetPackageInfoResponse{
	Name:    "requests",
	Version: "2.32.3",
	License: "Apache-2.0",
	DistributionPoints: []*oslcv1alpha.DistributionPoint{{
		Name:        "requests",
		Url:         "https://pypi.org/project/requests/",
		Distributor: oslc.DistributorPypi,
	}},
}
var npmTestEntry = oslc.Entry{
	Name:    "test",
	Version: "3.3.0",
	License: "MIT",
	DistributionPoints: []oslc.DistributionPoint{{
		Name:        "test",
		URL:         "https://www.npmjs.com/package/test",
		Distributor: oslc.DistributorNpm,
	}},
}

var npmTestGetPackageInfoRequest = oslcv1alpha.GetPackageInfoRequest{
	Name:        "test",
	Version:     "3.3.0",
	Distributor: oslc.DistributorNpm,
}

var npmTestGetPackageInfoResponse = oslcv1alpha.GetPackageInfoResponse{
	Name:    "test",
	Version: "3.3.0",
	License: "MIT",
	DistributionPoints: []*oslcv1alpha.DistributionPoint{{
		Name:        "test",
		Url:         "https://www.npmjs.com/package/test",
		Distributor: oslc.DistributorNpm,
	}},
}

var mavenLog4jEntry = oslc.Entry{
	Name:    "org.apache.logging.log4j:log4j",
	Version: "3.0.0-beta2",
	License: "Apache-2.0",
	DistributionPoints: []oslc.DistributionPoint{{
		Name:        "org.apache.logging.log4j:log4j",
		URL:         "https://central.sonatype.com/artifact/org.apache.logging.log4j/log4j",
		Distributor: oslc.DistributorMaven,
	}},
}

var mavenLog4jGetPackageInfoRequest = oslcv1alpha.GetPackageInfoRequest{
	Name:        "org.apache.logging.log4j:log4j",
	Version:     "3.0.0-beta2",
	Distributor: oslc.DistributorMaven,
}

var mavenLog4jGetPackageInfoResponse = oslcv1alpha.GetPackageInfoResponse{
	Name:    "org.apache.logging.log4j:log4j",
	Version: "3.0.0-beta2",
	License: "Apache-2.0",
	DistributionPoints: []*oslcv1alpha.DistributionPoint{{
		Name:        "org.apache.logging.log4j:log4j",
		Url:         "https://central.sonatype.com/artifact/org.apache.logging.log4j/log4j",
		Distributor: oslc.DistributorMaven,
	}},
}

var cratesIoSnarkVMEntry = oslc.Entry{
	Name:    "snarkvm-marlin",
	Version: "0.8.0",
	License: "GPL-3.0",
	DistributionPoints: []oslc.DistributionPoint{{
		Name:        "snarkvm-marlin",
		URL:         "https://crates.io/crates/snarkvm-marlin",
		Distributor: oslc.DistributorCratesIo,
	}},
}

var cratesIoSnarkVMGetPackageInfoRequest = oslcv1alpha.GetPackageInfoRequest{
	Name:        "snarkvm-marlin",
	Version:     "0.8.0",
	Distributor: oslc.DistributorCratesIo,
}

var cratesIoSnarkVMGetPackageInfoResponse = oslcv1alpha.GetPackageInfoResponse{
	Name:    "snarkvm-marlin",
	Version: "0.8.0",
	License: "GPL-3.0",
	DistributionPoints: []*oslcv1alpha.DistributionPoint{{
		Name:        "snarkvm-marlin",
		Url:         "https://crates.io/crates/snarkvm-marlin",
		Distributor: oslc.DistributorCratesIo,
	}},
}

func TestServer_GetPackageInfo(t *testing.T) {
	type fields struct {
		options *serverOptions
	}
	type args struct {
		ctx context.Context
		c   *oslcv1alpha.GetPackageInfoRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *oslcv1alpha.GetPackageInfoResponse
		wantErr bool
	}{
		{
			name: "datastore_has_entry",
			fields: fields{
				options: &serverOptions{
					Datastore: func() oslc.Datastore {
						mockDatastore := oslcMocks.NewMockDatastore(t)
						mockDatastore.EXPECT().Retrieve(context.Background(), pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version, oslc.DistributorPypi).
							Return(pypiRequestsEntry, nil)
						return mockDatastore
					}(),
				},
			},
			args: args{
				ctx: context.Background(),
				c:   &pypiRequestsGetPackageInfoRequest,
			},
			want:    &pypiRequestsGetPackageInfoResponse,
			wantErr: false,
		},
		{
			name: "datastore_no_entry",
			fields: fields{
				options: &serverOptions{
					Datastore: func() oslc.Datastore {
						mockDatastore := oslcMocks.NewMockDatastore(t)
						mockDatastore.EXPECT().Retrieve(context.Background(), pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version, oslc.DistributorPypi).
							Return(oslc.Entry{}, oslc.ErrDatastoreObjectNotFound)
						mockDatastore.EXPECT().Save(context.Background(), pypiRequestsEntry).
							Return(nil)
						return mockDatastore
					}(),
					PypiClient: func() oslc.DistributorClient {
						mockClient := oslcMocks.NewMockDistributorClient(t)
						mockClient.EXPECT().GetPackageVersion(pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version).
							Return(pypiRequestsEntry, nil)
						return mockClient
					}(),
					Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
					LicenseIDNormalizer: func() oslc.LicenseIDNormalizer {
						mockNormalizer := oslcMocks.NewMockLicenseIDNormalizer(t)
						mockNormalizer.EXPECT().NormalizeID(context.Background(), pypiRequestsEntry.License).
							Return(pypiRequestsEntry.License)
						return mockNormalizer
					}(),
				},
			},
			args: args{
				ctx: context.Background(),
				c:   &pypiRequestsGetPackageInfoRequest,
			},
			want:    &pypiRequestsGetPackageInfoResponse,
			wantErr: false,
		},
		{
			name: "datastore_retrieval_failure",
			fields: fields{
				options: &serverOptions{
					Datastore: func() oslc.Datastore {
						mockDatastore := oslcMocks.NewMockDatastore(t)
						mockDatastore.EXPECT().Retrieve(context.Background(), pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version, oslc.DistributorPypi).
							Return(oslc.Entry{}, assert.AnError)
						mockDatastore.EXPECT().Save(context.Background(), pypiRequestsEntry).
							Return(nil)
						return mockDatastore
					}(),
					Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
					PypiClient: func() oslc.DistributorClient {
						mockClient := oslcMocks.NewMockDistributorClient(t)
						mockClient.EXPECT().GetPackageVersion(pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version).
							Return(pypiRequestsEntry, nil)
						return mockClient
					}(),
					LicenseIDNormalizer: func() oslc.LicenseIDNormalizer {
						mockNormalizer := oslcMocks.NewMockLicenseIDNormalizer(t)
						mockNormalizer.EXPECT().NormalizeID(context.Background(), pypiRequestsEntry.License).
							Return(pypiRequestsEntry.License)
						return mockNormalizer
					}(),
				},
			},
			args: args{
				ctx: context.Background(),
				c:   &pypiRequestsGetPackageInfoRequest,
			},
			want:    &pypiRequestsGetPackageInfoResponse,
			wantErr: false,
		},
		{
			name: "datastore_save_failure",
			fields: fields{
				options: &serverOptions{
					Datastore: func() oslc.Datastore {
						mockDatastore := oslcMocks.NewMockDatastore(t)
						mockDatastore.EXPECT().Retrieve(context.Background(), pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version, oslc.DistributorPypi).
							Return(oslc.Entry{}, oslc.ErrDatastoreObjectNotFound)
						mockDatastore.EXPECT().Save(context.Background(), pypiRequestsEntry).
							Return(assert.AnError)
						return mockDatastore
					}(),
					Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
					PypiClient: func() oslc.DistributorClient {
						mockClient := oslcMocks.NewMockDistributorClient(t)
						mockClient.EXPECT().GetPackageVersion(pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version).
							Return(pypiRequestsEntry, nil)
						return mockClient
					}(),
					LicenseIDNormalizer: func() oslc.LicenseIDNormalizer {
						mockNormalizer := oslcMocks.NewMockLicenseIDNormalizer(t)
						mockNormalizer.EXPECT().NormalizeID(context.Background(), pypiRequestsEntry.License).
							Return(pypiRequestsEntry.License)
						return mockNormalizer
					}(),
				},
			},
			args: args{
				ctx: context.Background(),
				c:   &pypiRequestsGetPackageInfoRequest,
			},
			want:    &pypiRequestsGetPackageInfoResponse,
			wantErr: false,
		},
		{
			name: "upstream_distributor_client_failure",
			fields: fields{
				options: &serverOptions{
					Datastore: func() oslc.Datastore {
						mockDatastore := oslcMocks.NewMockDatastore(t)
						mockDatastore.EXPECT().Retrieve(context.Background(), pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version, oslc.DistributorPypi).
							Return(oslc.Entry{}, oslc.ErrDatastoreObjectNotFound)
						return mockDatastore
					}(),
					PypiClient: func() oslc.DistributorClient {
						mockClient := oslcMocks.NewMockDistributorClient(t)
						mockClient.EXPECT().GetPackageVersion(pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version).
							Return(oslc.Entry{}, assert.AnError)
						return mockClient
					}(),
					Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				},
			},
			args: args{
				ctx: context.Background(),
				c:   &pypiRequestsGetPackageInfoRequest,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "distributor_pypi_called",
			fields: fields{
				options: &serverOptions{
					Datastore: func() oslc.Datastore {
						mockDatastore := oslcMocks.NewMockDatastore(t)
						mockDatastore.EXPECT().Retrieve(context.Background(), pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version, oslc.DistributorPypi).
							Return(oslc.Entry{}, oslc.ErrDatastoreObjectNotFound)
						mockDatastore.EXPECT().Save(context.Background(), pypiRequestsEntry).
							Return(nil)
						return mockDatastore
					}(),
					PypiClient: func() oslc.DistributorClient {
						mockClient := oslcMocks.NewMockDistributorClient(t)
						mockClient.EXPECT().GetPackageVersion(pypiRequestsGetPackageInfoRequest.Name, pypiRequestsGetPackageInfoRequest.Version).
							Return(pypiRequestsEntry, nil)
						return mockClient
					}(),
					Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
					LicenseIDNormalizer: func() oslc.LicenseIDNormalizer {
						mockNormalizer := oslcMocks.NewMockLicenseIDNormalizer(t)
						mockNormalizer.EXPECT().NormalizeID(context.Background(), pypiRequestsEntry.License).
							Return(pypiRequestsEntry.License)
						return mockNormalizer
					}(),
				},
			},
			args: args{
				ctx: context.Background(),
				c:   &pypiRequestsGetPackageInfoRequest,
			},
			want:    &pypiRequestsGetPackageInfoResponse,
			wantErr: false,
		},
		{
			name: "distributor_npm_called",
			fields: fields{
				options: &serverOptions{
					Datastore: func() oslc.Datastore {
						mockDatastore := oslcMocks.NewMockDatastore(t)
						mockDatastore.EXPECT().Retrieve(context.Background(), npmTestGetPackageInfoRequest.Name, npmTestGetPackageInfoRequest.Version, oslc.DistributorNpm).
							Return(oslc.Entry{}, oslc.ErrDatastoreObjectNotFound)
						mockDatastore.EXPECT().Save(context.Background(), npmTestEntry).
							Return(nil)
						return mockDatastore
					}(),
					NpmClient: func() oslc.DistributorClient {
						mockClient := oslcMocks.NewMockDistributorClient(t)
						mockClient.EXPECT().GetPackageVersion(npmTestGetPackageInfoRequest.Name, npmTestGetPackageInfoRequest.Version).
							Return(npmTestEntry, nil)
						return mockClient
					}(),
					Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
					LicenseIDNormalizer: func() oslc.LicenseIDNormalizer {
						mockNormalizer := oslcMocks.NewMockLicenseIDNormalizer(t)
						mockNormalizer.EXPECT().NormalizeID(context.Background(), npmTestEntry.License).
							Return(npmTestEntry.License)
						return mockNormalizer
					}(),
				},
			},
			args: args{
				ctx: context.Background(),
				c:   &npmTestGetPackageInfoRequest,
			},
			want:    &npmTestGetPackageInfoResponse,
			wantErr: false,
		},
		{
			name: "distributor_maven_called",
			fields: fields{
				options: &serverOptions{
					Datastore: func() oslc.Datastore {
						mockDatastore := oslcMocks.NewMockDatastore(t)
						mockDatastore.EXPECT().Retrieve(context.Background(), mavenLog4jGetPackageInfoRequest.Name, mavenLog4jGetPackageInfoRequest.Version, oslc.DistributorMaven).
							Return(oslc.Entry{}, oslc.ErrDatastoreObjectNotFound)
						mockDatastore.EXPECT().Save(context.Background(), mavenLog4jEntry).
							Return(nil)
						return mockDatastore
					}(),
					MavenClient: func() oslc.DistributorClient {
						mockClient := oslcMocks.NewMockDistributorClient(t)
						mockClient.EXPECT().GetPackageVersion(mavenLog4jGetPackageInfoRequest.Name, mavenLog4jGetPackageInfoRequest.Version).
							Return(mavenLog4jEntry, nil)
						return mockClient
					}(),
					Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
					LicenseIDNormalizer: func() oslc.LicenseIDNormalizer {
						mockNormalizer := oslcMocks.NewMockLicenseIDNormalizer(t)
						mockNormalizer.EXPECT().NormalizeID(context.Background(), mavenLog4jEntry.License).
							Return(mavenLog4jEntry.License)
						return mockNormalizer
					}(),
				},
			},
			args: args{
				ctx: context.Background(),
				c:   &mavenLog4jGetPackageInfoRequest,
			},
			want:    &mavenLog4jGetPackageInfoResponse,
			wantErr: false,
		},
		{
			name: "distributor_cratesio_called",
			fields: fields{
				options: &serverOptions{
					Datastore: func() oslc.Datastore {
						mockDatastore := oslcMocks.NewMockDatastore(t)
						mockDatastore.EXPECT().Retrieve(context.Background(), cratesIoSnarkVMGetPackageInfoRequest.Name, cratesIoSnarkVMGetPackageInfoRequest.Version, oslc.DistributorCratesIo).
							Return(oslc.Entry{}, oslc.ErrDatastoreObjectNotFound)
						mockDatastore.EXPECT().Save(context.Background(), cratesIoSnarkVMEntry).
							Return(nil)
						return mockDatastore
					}(),
					CratesIoClient: func() oslc.DistributorClient {
						mockClient := oslcMocks.NewMockDistributorClient(t)
						mockClient.EXPECT().GetPackageVersion(cratesIoSnarkVMGetPackageInfoRequest.Name, cratesIoSnarkVMGetPackageInfoRequest.Version).
							Return(cratesIoSnarkVMEntry, nil)
						return mockClient
					}(),
					Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
					LicenseIDNormalizer: func() oslc.LicenseIDNormalizer {
						mockNormalizer := oslcMocks.NewMockLicenseIDNormalizer(t)
						mockNormalizer.EXPECT().NormalizeID(context.Background(), cratesIoSnarkVMEntry.License).
							Return(cratesIoSnarkVMEntry.License)
						return mockNormalizer
					}(),
				},
			},
			args: args{
				ctx: context.Background(),
				c:   &cratesIoSnarkVMGetPackageInfoRequest,
			},
			want:    &cratesIoSnarkVMGetPackageInfoResponse,
			wantErr: false,
		},
		{
			name: "invalid_distributor",
			fields: fields{
				options: &serverOptions{},
			},
			args: args{
				ctx: context.Background(),
				c: &oslcv1alpha.GetPackageInfoRequest{
					Distributor: "invalid",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Server{
				options: tt.fields.options,
			}
			got, err := s.GetPackageInfo(tt.args.ctx, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPackageInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServer_getPackageFromDistributor_invalid_distributor(t *testing.T) {
	s := Server{}
	_, err := s.getPackageFromDistributor(context.Background(), "invalid", "", "")
	var ide InvalidDistributorError
	require.ErrorAs(t, err, &ide)
}

func TestInvalidDistributorError_Error(t *testing.T) {
	ide := InvalidDistributorError{Distributor: "invalid"}
	require.Equal(t, "invalid distributor: invalid", ide.Error())
}
