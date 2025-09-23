package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/validate"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/auth"
	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/pb/pbconnect"
	"github.com/zanz1n/mc-manager/internal/server"
	"github.com/zanz1n/mc-manager/internal/utils"
	webstatic "github.com/zanz1n/mc-manager/web"
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
	distroRepo.AddDistribution(pb.Distribution_PAPER, distribution.NewPaper(nil))
	distroRepo.AddDistribution(pb.Distribution_VANILLA, distribution.NewVanilla(nil))

	runners := server.NewRunners(querier, nil)

	if cfg.LocalNode != nil && cfg.LocalNode.Enable {
		r, err := RunLocalNode(ctx, cfg.LocalNode, distroRepo, querier)
		if err != nil {
			log.Fatalln("Failed to run local node:", err)
		}
		runners.AddRunner(cfg.LocalNode.ID, r)
	}

	validator, err := validate.NewInterceptor()
	if err != nil {
		panic(err)
	}

	var localNodeId dto.Snowflake
	if cfg.LocalNode != nil {
		localNodeId = cfg.LocalNode.ID
	}

	interceptors := connect.WithInterceptors(
		utils.NewLoggerInterceptor(),
		utils.NewErrorInterceptor(),
		validator,
	)

	grpcR := chi.NewRouter()

	if !cfg.Standalone {
		grpcR.Use(stripPrefix("/api"))
	} else {
		grpcR.Use(cors.AllowAll().Handler)
		grpcR.Use(middleware.CleanPath)
	}
	grpcR.NotFound(http.NotFound)

	grpcR.Mount(pbconnect.NewAuthServiceHandler(
		server.NewAuthServer(querier, auther, authRepo, cfg),
		interceptors,
	))
	grpcR.Mount(pbconnect.NewUserServiceHandler(
		server.NewUserServer(querier, authRepo, cfg),
		interceptors,
	))
	grpcR.Mount(pbconnect.NewNodeServiceHandler(
		server.NewNodeServer(querier, authRepo, localNodeId),
		interceptors,
	))
	grpcR.Mount(pbconnect.NewInstanceServiceHandler(
		server.NewInstanceServer(querier, authRepo, runners),
		interceptors,
	))
	grpcR.Mount(pbconnect.NewDistributionServiceHandler(
		distribution.NewServer(distroRepo),
		interceptors,
	))

	if cfg.Server.EnableReflection {
		reflector := grpcreflect.NewStaticReflector(
			"manager.AuthService",
			"manager.UserService",
			"manager.NodeService",
			"manager.InstanceService",
			"manager.DistributionService",
		)
		grpcR.Mount(grpcreflect.NewHandlerV1(reflector, interceptors))
		grpcR.Mount(grpcreflect.NewHandlerV1Alpha(reflector, interceptors))
	}

	var r *chi.Mux
	if !cfg.Standalone {
		r = chi.NewRouter()
		r.Use(cors.AllowAll().Handler)
		r.Use(middleware.CleanPath)

		r.Mount("/", fileHandler(webstatic.Build))
		r.Mount("/api", grpcR)
	} else {
		r = grpcR
	}

	if err = utils.Serve(ctx, cfg.Server, r); err != nil {
		log.Fatalln(err)
	}
}

func stripPrefix(prefix string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, prefix)

			h.ServeHTTP(w, r)
		})
	}
}

func handleIndex(static fs.FS, w http.ResponseWriter, r *http.Request) {
	f, err := static.Open("build/index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	w.Header().Set("content-type", "text/html; charset=utf-8")
	io.Copy(w, f)
}

func fileHandler(static fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			handleIndex(static, w, r)
			return
		}

		f, err := static.Open(path.Join("build", r.URL.Path))
		if err != nil {
			handleIndex(static, w, r)
			return
		}
		defer f.Close()

		mimetype := mime.TypeByExtension(path.Ext(r.URL.Path))
		if mimetype == "" {
			mimetype = "text/plain"
		}

		w.Header().Set("content-type", mimetype)
		io.Copy(w, f)
	}
}
