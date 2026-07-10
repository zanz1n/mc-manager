package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/moby/moby/client"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/pb/pbconnect"
	"github.com/zanz1n/mc-manager/internal/runner"
)

func RunLocalNode(
	ctx context.Context,
	cfg *config.APILocalNodeConfig,
	distros *distribution.Repository,
	queries db.Querier,
) (pbconnect.RunnerServiceClient, error) {
	_, err := queries.NodeGetById(ctx, cfg.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		_, err = queries.NodeCreate(ctx, db.NodeCreateParams{
			ID:       cfg.ID,
			Name:     "Local Node",
			Endpoint: "passthrough://bufnet",
		})
	} else {
		_, err = queries.NodeUpdate(ctx, db.NodeUpdateParams{
			ID:       cfg.ID,
			Name:     "Local Node",
			Endpoint: "passthrough://bufnet",
		})
	}

	if err != nil {
		return nil, err
	}
	slog.Info("LocalNode: Created local node", "id", cfg.ID)

	start := time.Now()

	docker, err := client.New(client.FromEnv)
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

	ln, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		return nil, err
	}

	addr, err := netip.ParseAddrPort(ln.Addr().String())
	if err != nil {
		panic(err)
	}

	manager := runner.NewManager(runtime)
	runnerServer := runner.NewServer(manager, distros)

	mux := http.NewServeMux()
	mux.Handle(pbconnect.NewRunnerServiceHandler(runnerServer))

	httpsrv := http.Server{
		Addr:    ln.Addr().String(),
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	client := pbconnect.NewRunnerServiceClient(
		http.DefaultClient,
		fmt.Sprintf("http://127.0.0.1:%d", addr.Port()),
	)

	go httpsrv.Serve(ln)

	slog.Info(
		"LocalNode: Running local node",
		"took", time.Since(start).Round(time.Microsecond),
	)

	return client, nil
}
