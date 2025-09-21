package server

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/internal/auth"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/pb/pbconnect"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ pbconnect.InstanceServiceHandler = (*InstanceServer)(nil)

type InstanceServer struct {
	db db.Querier
	ar *auth.Respository
	r  *Runners

	pbconnect.UnimplementedInstanceServiceHandler
}

func NewInstanceServer(db db.Querier, ar *auth.Respository, r *Runners) *InstanceServer {
	return &InstanceServer{
		db: db,
		ar: ar,
		r:  r,
	}
}

// GetById implements pbconnect.InstanceServiceHandler.
func (s *InstanceServer) GetById(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[pb.Instance], error) {
	res := connect.NewResponse((*pb.Instance)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}
	id := dto.Snowflake(req.Msg.Id)

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

	snowflake := connect.NewRequest(&pb.Snowflake{Id: req.Msg.Id})
	ri, err := runner.GetStateById(ctx, snowflake)
	if err == nil {
		state, players = ri.Msg.State, ri.Msg.Players
	}

	res.Msg = i.IntoPB(state, players)
	return res, nil
}

// // GetMany implements pbconnect.InstanceServiceHandler.
// func (s *InstanceServer) GetMany(
// 	ctx context.Context,
// 	req *connect.Request[pb.Pagination],
// ) (*connect.Response[pb.InstanceGetManyResponse], error) {
// 	panic("unimplemented")
// }

// // GetByUser implements pbconnect.InstanceServiceHandler.
// func (s *InstanceServer) GetByUser(
// 	ctx context.Context,
// 	req *connect.Request[pb.InstanceGetByUserRequest],
// ) (*connect.Response[pb.InstanceGetManyResponse], error) {
// 	panic("unimplemented")
// }

// // GetByNode implements pbconnect.InstanceServiceHandler.
// func (s *InstanceServer) GetByNode(
// 	ctx context.Context,
// 	req *connect.Request[pb.InstanceGetByNodeRequest],
// ) (*connect.Response[pb.InstanceGetManyResponse], error) {
// 	panic("unimplemented")
// }

// Launch implements pbconnect.InstanceServiceHandler.
func (s *InstanceServer) Launch(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[emptypb.Empty], error) {
	res := connect.NewResponse((*emptypb.Empty)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}
	id := dto.Snowflake(req.Msg.Id)

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

	_, err = runner.Launch(ctx, connect.NewRequest(&pb.RunnerLaunchRequest{
		Id:            uint64(i.ID),
		Name:          i.Name,
		Version:       i.Version,
		VersionDistro: i.VersionDistro,
		Limits:        i.Limits,
		Config:        i.Config,
	}))
	if err != nil {
		return nil, err
	}

	if err = s.db.InstanceUpdateLastLaunched(ctx, id); err != nil {
		slog.Error(
			"InstanceServer: Failed to update `last_launched`",
			"id", id,
			"error", err,
		)
	}

	res.Msg = &emptypb.Empty{}
	return res, nil
}

// Stop implements pbconnect.InstanceServiceHandler.
func (s *InstanceServer) Stop(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[emptypb.Empty], error) {
	res := connect.NewResponse((*emptypb.Empty)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}
	id := dto.Snowflake(req.Msg.Id)

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

	snowflake := connect.NewRequest(&pb.Snowflake{Id: req.Msg.Id})
	_, err = runner.Stop(ctx, snowflake)
	if err != nil {
		return nil, err
	}

	res.Msg = &emptypb.Empty{}
	return res, nil
}

// Create implements pbconnect.InstanceServiceHandler.
func (s *InstanceServer) Create(
	ctx context.Context,
	req *connect.Request[pb.InstanceCreateRequest],
) (*connect.Response[pb.Instance], error) {
	res := connect.NewResponse((*pb.Instance)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}
	id := dto.NewSnowflake()

	i, err := s.db.InstanceCreate(ctx, db.InstanceCreateParams{
		ID:            id,
		UserID:        dto.Snowflake(req.Msg.UserId),
		NodeID:        dto.Snowflake(req.Msg.NodeId),
		Name:          req.Msg.Name,
		Description:   req.Msg.Description,
		Version:       req.Msg.Version,
		VersionDistro: req.Msg.VersionDistro,
		Config:        req.Msg.Config,
		Limits:        req.Msg.Limits,
	})
	if err != nil {
		return nil, err
	}

	res.Msg = i.IntoPB(pb.InstanceState_STATE_OFFLINE, 0)
	return res, nil
}

// SendCommand implements pbconnect.InstanceServiceHandler.
func (s *InstanceServer) SendCommand(
	ctx context.Context,
	req *connect.Request[pb.InstanceSendCommandRequest],
) (*connect.Response[emptypb.Empty], error) {
	res := connect.NewResponse((*emptypb.Empty)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	i, err := s.instanceGetById(ctx, dto.Snowflake(req.Msg.InstanceId))
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

	_, err = runner.SendCommand(ctx, connect.NewRequest(
		&pb.RunnerSendCommandRequest{
			InstanceId: req.Msg.InstanceId,
			Command:    req.Msg.Command,
		},
	))
	if err != nil {
		return nil, err
	}

	res.Msg = &emptypb.Empty{}
	return res, nil
}

// GetEvents implements pbconnect.InstanceServiceHandler.
func (s *InstanceServer) GetEvents(
	ctx context.Context,
	req *connect.Request[pb.InstanceGetEventsRequest],
	stream *connect.ServerStream[pb.Event],
) error {
	authed, err := s.ar.Authenticate(ctx, req.Header(), stream.ResponseHeader())
	if err != nil {
		return err
	}

	i, err := s.instanceGetById(ctx, dto.Snowflake(req.Msg.Id))
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

	nstream, err := runner.Listen(ctx, connect.NewRequest(
		&pb.RunnerListenRequest{
			InstanceId:  req.Msg.Id,
			IncludeLogs: req.Msg.IncludeLogs,
		},
	))
	if err != nil {
		return err
	}

	for {
		ok := nstream.Receive()
		if !ok {
			return nstream.Err()
		}

		event := nstream.Msg()

		if err = stream.Send(event); err != nil {
			return err
		}
	}
}

// Delete implements pbconnect.InstanceServiceHandler.
func (s *InstanceServer) Delete(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[pb.Instance], error) {
	res := connect.NewResponse((*pb.Instance)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	id := dto.Snowflake(req.Msg.Id)

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
			snowflake := connect.NewRequest(&pb.Snowflake{Id: req.Msg.Id})
			runner.Stop(ctx, snowflake)
		} else {
			slog.Error(
				"InstanceServer: Failed to call node to stop instance",
				"node_id", i.NodeID,
				"took", time.Since(start).Round(time.Millisecond),
				"error", err,
			)
		}
	}()

	res.Msg = i.IntoPB(pb.InstanceState_STATE_OFFLINE, 0)
	return res, nil
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
