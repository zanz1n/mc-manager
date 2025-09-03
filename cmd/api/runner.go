package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/docker/docker/client"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/runner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func RunLocalNode(
	ctx context.Context,
	cfg *config.APILocalNodeConfig,
	distros *distribution.Repository,
) (pb.RunnerServiceClient, error) {
	start := time.Now()

	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("connect to docker: %w", err)
	}

	slog.Info(
		"Docker: Client connected",
		"version", docker.ClientVersion(),
		"took", time.Since(start).Round(time.Microsecond),
	)

	runtime, err := runner.NewDockerRuntime(
		ctx,
		cfg.Docker,
		cfg.Data,
		docker,
		nil,
		runner.NewTemurinJre("noble"),
	)
	if err != nil {
		return nil, fmt.Errorf("create docker runner: %w", err)
	}

	manager := runner.NewManager(runtime)
	runnerServer := runner.NewServer(manager, distros)

	ln := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	pb.RegisterRunnerServiceServer(s, runnerServer)

	dialFn := func(ctx context.Context, s string) (net.Conn, error) {
		return ln.DialContext(ctx)
	}

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(dialFn),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := s.Serve(ln); err != nil {
			panic(err)
		}
	}()

	slog.Info(
		"LocalNode: Running local node",
		"took", time.Since(start).Round(time.Microsecond),
	)

	return pb.NewRunnerServiceClient(conn), nil
}
