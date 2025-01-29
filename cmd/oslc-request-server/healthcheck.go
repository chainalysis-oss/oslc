package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"strconv"
	"time"
)

var healthCheckCommand = &cli.Command{
	Name:   "healthcheck",
	Usage:  "Check the health of the grpc server",
	Action: healthcheckAction,
}

func healthcheckAction(cCtx *cli.Context) error {
	config, err := createConfiguration("config.json")
	if err != nil {
		return fmt.Errorf("failed to create configuration: %w", err)
	}
	logger := getLogger(config.Log.Level, config.Log.Kind)

	conn, err := grpc.NewClient(
		net.JoinHostPort(config.Grpc.Interface, strconv.Itoa(config.Grpc.Port)),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})),
	)
	if err != nil {
		return fmt.Errorf("failed to create grpc client: %w", err)
	}
	healthClient := healthgrpc.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := healthClient.Check(ctx, &healthgrpc.HealthCheckRequest{
		Service: "",
	})
	if err != nil {
		return fmt.Errorf("failed to check health: %w", err)
	}
	if resp.Status != healthgrpc.HealthCheckResponse_SERVING {
		return fmt.Errorf("service is not serving")
	}
	logger.Info("service is serving")
	return nil
}
