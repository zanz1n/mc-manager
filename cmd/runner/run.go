package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"time"

	"buf.build/go/protovalidate"
	"github.com/docker/docker/client"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/runner"
	"github.com/zanz1n/mc-manager/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Run(ctx context.Context, cfg *config.Config) {
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
		cfg,
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
	cfg *config.Config,
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

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			utils.LoggerUnaryServerInterceptor,
			utils.ErrorUnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			utils.LoggerStreamServerInterceptor,
			utils.ErrorStreamServerInterceptor,
		),
	}

	if cfg.Server.Password != "" {
		opts = append(opts,
			grpc.ChainUnaryInterceptor(
				utils.AuthUnaryServerInterceptor(cfg.Server.Password),
			),
			grpc.ChainStreamInterceptor(
				utils.AuthStreamServerInterceptor(cfg.Server.Password),
			),
		)
	}

	validator, err := protovalidate.New()
	if err != nil {
		panic(err)
	}

	opts = append(opts, grpc.ChainUnaryInterceptor(
		protovalidate_middleware.UnaryServerInterceptor(validator),
	))

	instanceServer := runner.NewServer(manager, distributions)

	server := grpc.NewServer(opts...)
	pb.RegisterDistributionServiceServer(
		server,
		distribution.NewServer(distributions),
	)
	pb.RegisterRunnerServiceServer(server, instanceServer)

	if cfg.Server.EnableReflection {
		reflection.Register(server)
	}

	go server.Serve(ln)
	defer func() {
		start := time.Now()
		graceful := utils.CloseGrpc(3*time.Second, server)
		slog.Info(
			"GRPC: Closed server",
			"graceful", graceful,
			"took", time.Since(start).Round(time.Millisecond),
		)
	}()

	<-ctx.Done()
}
