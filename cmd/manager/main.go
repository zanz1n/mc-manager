package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/client"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/instance"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String(
	"config",
	"/etc/mc/config.yaml",
	"the path of the config file",
)

var endCh = make(chan os.Signal, 1)

func init() {
	flag.Parse()

	if cenv := os.Getenv("CONFIG_FILE"); cenv != "" {
		*configFile = cenv
	}

	signal.Notify(endCh, syscall.SIGINT, syscall.SIGTERM)
}

func main() {
	cfg, err := config.GetConfig(*configFile)
	if err != nil {
		log.Fatalln("Failed to get config:", err)
	}

	err = config.WriteConfig(*configFile, cfg)
	if err != nil {
		log.Fatalln("Failed to format config:", err)
	}

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

	runner, err := instance.NewDockerRunner(
		context.Background(),
		cfg,
		docker,
		nil,
		instance.NewTemurinJre("noble"),
	)
	if err != nil {
		log.Fatalln("Failed to create docker runner:", err)
	}

	manager := instance.NewManager(runner)

	start = time.Now()
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

	server := grpc.NewServer(opts...)
	pb.RegisterDistributionServiceServer(
		server,
		distribution.NewServer(distributions),
	)
	pb.RegisterInstanceServiceServer(
		server,
		instance.NewServer(manager, distributions),
	)
	if cfg.Server.EnableReflection {
		reflection.Register(server)
	}

	go server.Serve(ln)
	defer func() {
		start := time.Now()
		server.GracefulStop()
		slog.Info(
			"GRPC: Closed server",
			"took", time.Since(start).Round(time.Millisecond),
		)
	}()

	sig := <-endCh
	log.Printf("Received signal %s: closing server ...\n", sig.String())
}
