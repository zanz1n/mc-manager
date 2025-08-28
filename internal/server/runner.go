package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/utils"
	"google.golang.org/grpc"
)

type runners struct {
	db db.Querier

	m  map[dto.Snowflake]pb.RunnerServiceClient
	mu sync.Mutex
}

func (r *runners) Get(ctx context.Context, id dto.Snowflake) (pb.RunnerServiceClient, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.m[id]
	if ok {
		return s, nil
	}
	return r.getSlow(ctx, id)
}

func (r *runners) getSlow(ctx context.Context, id dto.Snowflake) (pb.RunnerServiceClient, error) {
	node, err := r.db.NodeGetById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrNodeNotFound, errors.New(id.String()))
		}
		return nil, err
	}

	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", node.Endpoint, node.GrpcPort),
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
		return nil, errors.Join(ErrNodeUnreachable, err)
	}

	s := pb.NewRunnerServiceClient(conn)
	r.m[id] = s

	return s, nil
}
