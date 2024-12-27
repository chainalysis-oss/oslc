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
	"fmt"
	oslcv1alpha "github.com/chainalysis-oss/oslc/gen/oslc/oslc/v1alpha"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net"
	"os"
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
		},
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

	host, err := rs.Host(ctx)
	require.NoError(t, err)

	port, err := rs.MappedPort(ctx, "8080")
	require.NoError(t, err)

	conn, err := grpc.NewClient(net.JoinHostPort(host, port.Port()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := oslcv1alpha.NewOslcServiceClient(conn)

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
}
