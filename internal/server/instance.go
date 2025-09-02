package server

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/zanz1n/mc-manager/internal/auth"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ pb.InstanceServiceServer = (*InstanceServer)(nil)

type InstanceServer struct {
	db db.Querier
	ar *auth.Respository
	r  runners

	pb.UnimplementedInstanceServiceServer
}

func NewInstanceServer(db db.Querier, ar *auth.Respository) *InstanceServer {
	return &InstanceServer{
		db: db,
		ar: ar,
		r: runners{
			db: db,
			m:  make(map[dto.Snowflake]pb.RunnerServiceClient),
		},
	}
}

// GetById implements pb.InstanceServiceServer.
func (s *InstanceServer) GetById(
	ctx context.Context,
	req *pb.Snowflake,
) (*pb.Instance, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	id := dto.Snowflake(req.Id)

	i, err := s.instanceGetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		if authed.GetId() != i.UserID {
			return nil, ErrPermissionDenied
		}
	}

	runner, err := s.r.Get(ctx, i.NodeID)
	if err != nil {
		return nil, err
	}

	state, players := pb.InstanceState_STATE_OFFLINE, int32(0)
	ri, err := runner.GetStateById(ctx, &pb.Snowflake{Id: req.Id})
	if err == nil {
		state, players = ri.State, ri.Players
	}

	return i.IntoPB(state, players), nil
}

// GetMany implements pb.InstanceServiceServer.
// func (r *InstanceServer) GetMany(
// 	ctx context.Context,
// 	req *pb.Pagination,
// ) (*pb.InstanceGetManyResponse, error) {
// 	panic("unimplemented")
// }

// GetByUser implements pb.InstanceServiceServer.
// func (r *InstanceServer) GetByUser(
// 	ctx context.Context,
// 	req *pb.InstanceGetByUserRequest,
// ) (*pb.InstanceGetManyResponse, error) {
// 	panic("unimplemented")
// }

// GetByNode implements pb.InstanceServiceServer.
// func (r *InstanceServer) GetByNode(
// 	ctx context.Context,
// 	req *pb.InstanceGetByNodeRequest,
// ) (*pb.InstanceGetManyResponse, error) {
// 	panic("unimplemented")
// }

// Launch implements pb.InstanceServiceServer.
func (s *InstanceServer) Launch(
	ctx context.Context,
	req *pb.Snowflake,
) (*emptypb.Empty, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	id := dto.Snowflake(req.Id)

	i, err := s.instanceGetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		if authed.GetId() != i.UserID {
			return nil, ErrPermissionDenied
		}
	}

	runner, err := s.r.Get(ctx, i.NodeID)
	if err != nil {
		return nil, err
	}

	_, err = runner.Launch(ctx, &pb.RunnerLaunchRequest{
		Id:            uint64(i.ID),
		Name:          i.Name,
		Version:       i.Version,
		VersionDistro: i.VersionDistro,
		Limits:        i.Limits,
		Config:        i.Config,
	})
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Create implements pb.InstanceServiceServer.
func (s *InstanceServer) Create(
	ctx context.Context,
	req *pb.InstanceCreateRequest,
) (*pb.Instance, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	id := dto.NewSnowflake()

	i, err := s.db.InstanceCreate(ctx, db.InstanceCreateParams{
		ID:            id,
		UserID:        dto.Snowflake(req.UserId),
		NodeID:        dto.Snowflake(req.NodeId),
		Name:          req.Name,
		Description:   req.Description,
		Version:       req.Version,
		VersionDistro: req.VersionDistro,
		Config:        req.Config,
		Limits:        req.Limits,
	})
	if err != nil {
		return nil, err
	}

	return i.IntoPB(pb.InstanceState_STATE_OFFLINE, 0), nil
}

// SendCommand implements pb.InstanceServiceServer.
func (s *InstanceServer) SendCommand(
	ctx context.Context,
	req *pb.InstanceSendCommandRequest,
) (*emptypb.Empty, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	i, err := s.instanceGetById(ctx, dto.Snowflake(req.InstanceId))
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		if authed.GetId() != i.UserID {
			return nil, ErrPermissionDenied
		}
	}

	runner, err := s.r.Get(ctx, i.NodeID)
	if err != nil {
		return nil, err
	}

	_, err = runner.SendCommand(ctx, &pb.RunnerSendCommandRequest{
		InstanceId: req.InstanceId,
		Command:    req.Command,
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// GetEvents implements pb.InstanceServiceServer.
func (s *InstanceServer) GetEvents(
	req *pb.InstanceGetEventsRequest,
	stream grpc.ServerStreamingServer[pb.Event],
) error {
	ctx := stream.Context()

	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return err
	}

	i, err := s.instanceGetById(ctx, dto.Snowflake(req.Id))
	if err != nil {
		return err
	}

	if !authed.IsAdmin() {
		if authed.GetId() != i.UserID {
			return ErrPermissionDenied
		}
	}

	runner, err := s.r.Get(ctx, i.NodeID)
	if err != nil {
		return err
	}

	res, err := runner.Listen(ctx, &pb.RunnerListenRequest{
		InstanceId:  req.Id,
		IncludeLogs: req.IncludeLogs,
	})
	if err != nil {
		return err
	}

	for {
		event, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if err = stream.Send(event); err != nil {
			return err
		}
	}
}

// Delete implements pb.InstanceServiceServer.
func (s *InstanceServer) Delete(
	ctx context.Context,
	req *pb.Snowflake,
) (*pb.Instance, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	id := dto.Snowflake(req.Id)

	i, err := s.db.InstanceDelete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrInstanceNotFound, errors.New(id.String()))
		}
		return nil, err
	}

	go func() {
		start := time.Now()

		runner, err := s.r.Get(ctx, i.NodeID)
		if err == nil {
			runner.Stop(ctx, &pb.Snowflake{Id: req.Id})
		} else {
			slog.Error(
				"InstanceServer: Failed to call node to stop instance",
				"node_id", i.NodeID,
				"took", time.Since(start).Round(time.Millisecond),
				"error", err,
			)
		}
	}()

	return i.IntoPB(pb.InstanceState_STATE_OFFLINE, 0), nil
}

func (s *InstanceServer) instanceGetById(
	ctx context.Context,
	id dto.Snowflake,
) (db.Instance, error) {
	i, err := s.db.InstanceGetById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrInstanceNotFound, errors.New(id.String()))
		}
	}

	return i, err
}
