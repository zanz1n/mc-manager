package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Runners struct {
	db db.Querier

	m  map[dto.Snowflake]pb.RunnerServiceClient
	mu sync.Mutex
}

func NewRunners(db db.Querier) *Runners {
	return &Runners{
		db: db,
		m:  make(map[dto.Snowflake]pb.RunnerServiceClient),
	}
}

func (r *Runners) AddRunner(id dto.Snowflake, s pb.RunnerServiceClient) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.m[id] = s
}

func (r *Runners) Get(ctx context.Context, id dto.Snowflake) (pb.RunnerServiceClient, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.m[id]
	if ok {
		return s, nil
	}
	return r.getSlow(ctx, id)
}

func (r *Runners) getSlow(ctx context.Context, id dto.Snowflake) (pb.RunnerServiceClient, error) {
	node, err := r.db.NodeGetById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrNodeNotFound, errors.New(id.String()))
		}
		return nil, err
	}

	start := time.Now()

	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", node.Endpoint, node.GrpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			utils.LoggerUnaryClientInterceptor,
			utils.AuthUnaryClientInterceptor(node.Token),
		),
		grpc.WithChainStreamInterceptor(
			utils.LoggerStreamClientInterceptor,
			utils.AuthStreamClientInterceptor(node.Token),
		),
	)
	if err != nil {
		slog.Error(
			"InstanceServer: Failed to reach node",
			"id", id,
			"took", time.Since(start).Round(time.Microsecond),
			"error", err,
		)
		return nil, errors.Join(ErrNodeUnreachable, err)
	}
	slog.Info(
		"InstanceServer: Connected to node",
		"id", id,
		"took", time.Since(start).Round(time.Microsecond),
	)

	s := pb.NewRunnerServiceClient(conn)
	r.m[id] = s

	return s, nil
}
