//go:build acceptance

// Note for running on macOS with Podman:
// When running these tests, the following 2 prefixes are required for testcontainers to work with Podman:
//  1. DOCKER_HOST environment variable must be set.
//     The format of the value should be `unix://<your_podman_socket_location>`. The podman socket location can be found
//     by running `podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}'`.
//     Alternatively, `/run/podman/podman.sock` (obtained from `podman info --format '{{.Host.RemoteSocket.Path}}'`) is
//     known to work.
//  2. TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED environment variable must be set to true.
package acceptance

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	oslcv1alpha "github.com/chainalysis-oss/oslc/gen/oslc/oslc/v1alpha"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

import (
	"context"
)

var ImageToTest string

func init() {
	var found bool
	ImageToTest, found = os.LookupEnv("IMAGE_TO_TEST")
	if !found {
		panic("IMAGE_TO_TEST environment variable must be set")
	}
}

func setupGrpcClientForOSLC(t *testing.T, ctx context.Context, c testcontainers.Container) *grpc.ClientConn {
	t.Helper()
	host, err := c.Host(ctx)
	require.NoError(t, err)

	port, err := c.MappedPort(ctx, "8080")
	require.NoError(t, err)

	addr := net.JoinHostPort(host, port.Port())

	caCert, err := os.ReadFile("../../build/tls/ca/rootCA.pem")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{RootCAs: caCertPool})),
	)
	require.NoError(t, err)
	return conn
}

func TestContainerStarts(t *testing.T) {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)
	pgContainerHost, err := postgresContainer.ContainerIP(ctx)
	require.NoError(t, err)

	certPath, err := filepath.Abs(filepath.Join("..", "..", "build", "tls", "oslc-request-server.internal.crt"))
	require.NoError(t, err)
	keyPath, err := filepath.Abs(filepath.Join("..", "..", "build", "tls", "oslc-request-server.internal.key"))

	req := testcontainers.ContainerRequest{
		Image:        ImageToTest,
		ExposedPorts: []string{"8080/tcp"},
		Entrypoint:   []string{"/usr/bin/oslc-request-server"},
		Env: map[string]string{
			"DATASTORE_HOST":     pgContainerHost,
			"DATASTORE_PORT":     "5432",
			"DATASTORE_DB":       "test",
			"DATASTORE_USER":     "user",
			"DATASTORE_PASSWORD": "password",
			"OSLC_TLS_CERT_FILE": "/oslc_tls_cert_file",
			"OSLC_TLS_KEY_FILE":  "/oslc_tls_key_file",
		},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      certPath,
				ContainerFilePath: "/oslc_tls_cert_file",
				FileMode:          0o777,
			},
			{
				HostFilePath:      keyPath,
				ContainerFilePath: "/oslc_tls_key_file",
				FileMode:          0o777,
			},
		},
		WaitingFor: wait.ForLog("starting grpc server").WithStartupTimeout(5 * time.Second),
	}
	rs, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	defer testcontainers.CleanupContainer(t, rs)
	defer testcontainers.CleanupContainer(t, postgresContainer)

	clogs, err := rs.Logs(ctx)
	require.NoError(t, err)
	readLogs, err := io.ReadAll(clogs)
	require.NoError(t, err)
	fmt.Printf("container logs: \n%s", string(readLogs))

	require.True(t, rs.IsRunning(), "container is not running")

	client := oslcv1alpha.NewOslcServiceClient(setupGrpcClientForOSLC(t, ctx, rs))

	resp, err := client.GetPackageInfo(ctx, &oslcv1alpha.GetPackageInfoRequest{Name: "requests", Version: "2.32.0", Distributor: "pypi"})
	require.NoError(t, err)
	require.EqualExportedValues(t, &oslcv1alpha.GetPackageInfoResponse{
		Name:    "requests",
		Version: "2.32.0",
		License: "Apache-2.0",
		DistributionPoints: []*oslcv1alpha.DistributionPoint{
			{
				Name:        "requests",
				Url:         "https://pypi.org/project/requests/",
				Distributor: "pypi",
			},
		},
	}, resp)

	resp, err = client.GetPackageInfo(ctx, &oslcv1alpha.GetPackageInfoRequest{Name: "github.com/keltia/leftpad", Version: "v0.1.0", Distributor: "go"})
	require.NoError(t, err)
	require.EqualExportedValues(t, &oslcv1alpha.GetPackageInfoResponse{
		Name:    "github.com/keltia/leftpad",
		Version: "v0.1.0",
		License: "BSD-2-Clause",
		DistributionPoints: []*oslcv1alpha.DistributionPoint{
			{
				Name:        "github.com/keltia/leftpad",
				Url:         "https://proxy.golang.org/github.com/keltia/leftpad/@v/v0.1.0.zip",
				Distributor: "go",
			},
		},
	}, resp)
}
