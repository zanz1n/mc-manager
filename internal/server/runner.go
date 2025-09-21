package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb/pbconnect"
	"github.com/zanz1n/mc-manager/internal/utils"
)

type Runners struct {
	db db.Querier
	c  *http.Client

	m  map[dto.Snowflake]pbconnect.RunnerServiceClient
	mu sync.Mutex
}

func NewRunners(db db.Querier, c *http.Client) *Runners {
	if c == nil {
		c = http.DefaultClient
	}

	return &Runners{
		db: db,
		m:  make(map[dto.Snowflake]pbconnect.RunnerServiceClient),
	}
}

func (r *Runners) AddRunner(id dto.Snowflake, s pbconnect.RunnerServiceClient) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.m[id] = s
}

func (r *Runners) Get(ctx context.Context, id dto.Snowflake) (pbconnect.RunnerServiceClient, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.m[id]
	if ok {
		return s, nil
	}
	return r.getSlow(ctx, id)
}

func (r *Runners) getSlow(ctx context.Context, id dto.Snowflake) (pbconnect.RunnerServiceClient, error) {
	node, err := r.db.NodeGetById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrNodeNotFound, errors.New(id.String()))
		}
		return nil, err
	}

	var urlSchema string
	if node.EndpointTls {
		urlSchema = "https"
	} else {
		urlSchema = "http"
	}

	baseUrl := fmt.Sprintf("%s://%s:%d", urlSchema, node.Endpoint, node.GrpcPort)

	start := time.Now()

	runner := pbconnect.NewRunnerServiceClient(
		r.c,
		baseUrl,
		connect.WithInterceptors(
			utils.NewLoggerInterceptor(),
			utils.NewAuthInterceptor(node.Token),
		),
	)

	slog.Info(
		"InstanceServer: Connected to node",
		"id", id,
		"took", time.Since(start).Round(time.Microsecond),
	)

	return runner, nil
}
