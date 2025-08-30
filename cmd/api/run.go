package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	"buf.build/go/protovalidate"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/auth"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/server"
	"github.com/zanz1n/mc-manager/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Run(ctx context.Context, cfg *config.APIConfig) {
	pub, priv, err := loadEd25519(cfg)
	if err != nil {
		log.Fatalln("Failed to load Ed25519 key pair:", err)
	}

	start := time.Now()
	querier, sqldb, err := openDB(ctx, cfg)
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}
	defer func() {
		start := time.Now()
		err := sqldb.Close()
		slog.Info(
			"DB: Closed client",
			"error", err,
			"took", time.Since(start).Round(time.Microsecond),
		)
	}()

	dbStats := sqldb.Stats()
	slog.Info(
		"DB: Client connected",
		"conns", fmt.Sprintf("%d/%d",
			dbStats.OpenConnections,
			dbStats.MaxOpenConnections,
		),
		"took", time.Since(start).Round(time.Microsecond),
	)

	start = time.Now()
	kvstorer, err := openKV(ctx, cfg)
	if err != nil {
		log.Fatalln("Failed to connect to valkey:", err)
	}
	defer func() {
		start := time.Now()
		err := kvstorer.Close()
		slog.Info(
			"Redis: Closed client",
			"error", err,
			"took", time.Since(start).Round(time.Microsecond),
		)
	}()

	slog.Info(
		"Redis: Client connected",
		"took", time.Since(start).Round(time.Microsecond),
	)

	auther := auth.NewJWTAuther(kvstorer, priv, pub, cfg)
	authRepo := auth.NewRepository(auther, querier, cfg)

	distroRepo := distribution.NewRepository()
	distroRepo.AddDistribution(
		pb.Distribution_PAPER,
		distribution.NewPaper(nil),
	)
	distroRepo.AddDistribution(
		pb.Distribution_VANILLA,
		distribution.NewVanilla(nil),
	)

	Serve(
		ctx,
		cfg,
		querier,
		auther,
		authRepo,
		distroRepo,
	)
}

func Serve(
	ctx context.Context,
	cfg *config.APIConfig,
	querier db.Querier,
	auther auth.Auther,
	authRepo *auth.Respository,
	distroRepo *distribution.Repository,
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

	validator, err := protovalidate.New()
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			utils.LoggerUnaryServerInterceptor,
			utils.ErrorUnaryServerInterceptor,
			protovalidate_middleware.UnaryServerInterceptor(validator),
		),
		grpc.ChainStreamInterceptor(
			utils.LoggerStreamServerInterceptor,
			utils.ErrorStreamServerInterceptor,
		),
	)

	pb.RegisterAuthServiceServer(
		grpcServer,
		server.NewAuthServer(querier, auther, authRepo, cfg),
	)
	pb.RegisterUserServiceServer(
		grpcServer,
		server.NewUserServer(querier, authRepo, cfg),
	)
	pb.RegisterInstanceServiceServer(
		grpcServer,
		server.NewInstanceServer(querier, authRepo),
	)
	pb.RegisterDistributionServiceServer(
		grpcServer,
		distribution.NewServer(distroRepo),
	)

	if cfg.Server.EnableReflection {
		reflection.Register(grpcServer)
	}

	go grpcServer.Serve(ln)
	defer func() {
		start := time.Now()
		graceful := utils.CloseGrpc(3*time.Second, grpcServer)
		slog.Info(
			"GRPC: Closed server",
			"graceful", graceful,
			"took", time.Since(start).Round(time.Millisecond),
		)
	}()

	<-ctx.Done()
}
