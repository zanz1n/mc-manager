package main

import (
	"context"
	"log"
	"log/slog"
	"net"
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

func Run(ctx context.Context, cfg *config.RunnerConfig) {
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
	distributions.AddDistribution(
		pb.Distribution_VANILLA,
		distribution.NewVanilla(nil),
	)
	distributions.AddDistribution(
		pb.Distribution_PAPER,
		distribution.NewPaper(nil),
	)

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

	Serve(ctx, cfg, distributions, manager)
}

func Serve(
	ctx context.Context,
	cfg *config.RunnerConfig,
	distributions *distribution.Repository,
	manager *runner.Manager,
) {
	start := time.Now()
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   cfg.Server.IP,
		Port: int(cfg.Server.Port),
	})
	if err != nil {
		log.Fatalln("Failed to listen tcp:", err)
	}

	slog.Info(
		"GRPC: Listening",
		"addr", ln.Addr(),
		"took", time.Since(start).Round(time.Microsecond),
	)

	validator, err := validate.NewInterceptor()
	if err != nil {
		panic(err)
	}

	loggerInterceptor := utils.NewLoggerInterceptor()
	errorInterceptor := utils.NewErrorInterceptor()

	mux := http.NewServeMux()

	mux.Handle(pbconnect.NewRunnerServiceHandler(
		runner.NewServer(manager, distributions),
		connect.WithInterceptors(loggerInterceptor, errorInterceptor, validator),
	))

	if cfg.Server.EnableReflection {
		reflector := grpcreflect.NewStaticReflector("manager.RunnerService")
		mux.Handle(grpcreflect.NewHandlerV1(reflector))
		mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	}

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	s := &http.Server{
		Addr:      ln.Addr().String(),
		Handler:   mux,
		Protocols: protocols,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		<-ctx.Done()

		shutdownStart := time.Now()

		stopctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := s.Shutdown(stopctx)
		slog.Info(
			"GRPC: Closed server",
			"took", time.Since(shutdownStart).Round(time.Millisecond),
			"online_for", time.Since(start).Round(time.Second),
			"error", err,
		)
	}()

	if err = s.Serve(ln); err != nil {
		log.Fatalln("Failed to grpc serve:", err)
	}
}
