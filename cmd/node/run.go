package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/validate"
	"github.com/docker/docker/client"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/pb/pbconnect"
	"github.com/zanz1n/mc-manager/internal/runner"
	"github.com/zanz1n/mc-manager/internal/utils"
)

func Run(ctx context.Context, cfg *config.NodeConfig) {
	start := time.Now()
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalln("Failed to connect to docker:", err)
	}

	defer func() {
		start := time.Now()
		err := docker.Close()
		slog.Info(
			"Docker: Closed client",
			"error", err,
			"took", time.Since(start).Round(time.Microsecond),
		)
	}()

	slog.Info(
		"Docker: Client connected",
		"version", docker.ClientVersion(),
		"took", time.Since(start).Round(time.Microsecond),
	)

	distributions := distribution.NewRepository()
	distributions.AddDistribution(pb.Distribution_VANILLA, distribution.NewVanilla(nil))
	distributions.AddDistribution(pb.Distribution_PAPER, distribution.NewPaper(nil))

	runtime, err := runner.NewDockerRuntime(
		context.Background(),
		cfg.Docker,
		cfg.Data,
		docker,
		nil,
		runner.NewTemurinJre("noble"),
	)
	if err != nil {
		log.Fatalln("Failed to create docker runner:", err)
	}

	manager := runner.NewManager(runtime)

	interceptors := connect.WithInterceptors(
		utils.NewLoggerInterceptor(),
		utils.NewErrorInterceptor(),
		validate.NewInterceptor(),
	)

	mux := http.NewServeMux()

	mux.Handle(pbconnect.NewRunnerServiceHandler(
		runner.NewServer(manager, distributions),
		interceptors,
	))

	if cfg.Server.EnableReflection {
		reflector := grpcreflect.NewStaticReflector("manager.RunnerService")
		mux.Handle(grpcreflect.NewHandlerV1(reflector, interceptors))
		mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector, interceptors))
	}

	if err = utils.Serve(ctx, cfg.Server, mux); err != nil {
		log.Fatalln(err)
	}
}
