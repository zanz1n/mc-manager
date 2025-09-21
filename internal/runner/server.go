package runner

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/pb/pbconnect"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ pbconnect.RunnerServiceHandler = (*Server)(nil)

type Server struct {
	m        *Manager
	versions *distribution.Repository
	pbconnect.UnimplementedRunnerServiceHandler
}

func NewServer(m *Manager, v *distribution.Repository) *Server {
	return &Server{m: m, versions: v}
}

// GetById implements pbconnect.RunnerServiceHandler.
func (s *Server) GetById(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[pb.RunningInstance], error) {
	i, err := s.m.GetById(ctx, dto.Snowflake(req.Msg.Id))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(i.IntoPB()), nil
}

// GetStateById implements pbconnect.RunnerServiceHandler.
func (s *Server) GetStateById(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[pb.RunnerGetStateResponse], error) {
	i, err := s.m.GetById(ctx, dto.Snowflake(req.Msg.Id))
	if err != nil {
		return nil, err
	}

	var players int32 = 0
	if i.proxy != nil {
		players = i.proxy.Players.Load()
	}

	return connect.NewResponse(&pb.RunnerGetStateResponse{
		Players: players,
		State:   i.GetState(),
	}), nil
}

// Launch implements pbconnect.RunnerServiceHandler.
func (s *Server) Launch(
	ctx context.Context,
	req *connect.Request[pb.RunnerLaunchRequest],
) (*connect.Response[pb.RunningInstance], error) {
	var (
		version distribution.Version
		err     error
	)
	if req.Msg.Version == "" {
		version, err = s.versions.GetLatest(ctx, req.Msg.VersionDistro)
	} else {
		version, err = s.versions.GetVersion(ctx, req.Msg.VersionDistro, req.Msg.Version)
	}

	if err != nil {
		return nil, err
	}

	var (
		limits InstanceLimits
		config InstanceConfig
	)
	limits.FromPB(req.Msg.Limits)
	config.FromPB(req.Msg.Config)

	i, err := s.m.Launch(ctx, InstanceCreateData{
		ID:      dto.Snowflake(req.Msg.Id),
		Name:    req.Msg.Name,
		Version: version,
		Limits:  limits,
		Config:  config,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(i.IntoPB()), nil
}

// Stop implements pbconnect.RunnerServiceHandler.
func (s *Server) Stop(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[pb.RunningInstance], error) {
	i, err := s.m.GetById(ctx, dto.Snowflake(req.Msg.Id))
	if err != nil {
		return nil, err
	}

	// TODO: timeout
	err = s.m.Stop(ctx, dto.Snowflake(req.Msg.Id))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(i.IntoPB()), nil
}

// SendCommand implements pbconnect.RunnerServiceHandler.
func (s *Server) SendCommand(
	ctx context.Context,
	req *connect.Request[pb.RunnerSendCommandRequest],
) (*connect.Response[emptypb.Empty], error) {
	i, err := s.m.GetById(ctx, dto.Snowflake(req.Msg.InstanceId))
	if err != nil {
		return nil, err
	}

	if err = i.SendCommand(req.Msg.Command); err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// Listen implements pbconnect.RunnerServiceHandler.
func (s *Server) Listen(
	ctx context.Context,
	req *connect.Request[pb.RunnerListenRequest],
	stream *connect.ServerStream[pb.Event],
) error {
	i, err := s.m.GetById(ctx, dto.Snowflake(req.Msg.InstanceId))
	if err != nil {
		return err
	}

	ch := i.AttachListener(req.Msg.IncludeLogs)
	defer func() {
		if !i.DetachListener(ch) {
			slog.Error(
				"Instance: Failed to detach listener",
				"id", i.ID,
			)
		}
	}()

	headers := stream.ResponseHeader()

	headers.Set("X-Instance-Listeners", strconv.Itoa(i.ListenersCount()))
	headers.Set(
		"X-Instance-Uptime",
		time.Since(i.LaunchedAt).Round(time.Second).String(),
	)

	var players int32 = 0
	if i.proxy != nil {
		players = i.proxy.Players.Load()
	}
	headers.Set("X-Instance-Players", strconv.Itoa(int(players)))

	for {
		ev, ok := <-ch
		if !ok {
			break
		}
		if err = stream.Send(ev.IntoPB()); err != nil {
			return err
		}
	}

	return nil
}

// ListenMany implements pbconnect.RunnerServiceHandler.
func (s *Server) ListenMany(
	ctx context.Context,
	req *connect.Request[pb.RunnerListenManyRequest],
	stream *connect.ServerStream[pb.RunnerListenManyResponse],
) error {
	instances, err := s.m.GetMany(ctx, req.Msg.Instances)
	if err != nil {
		return err
	}

	type evt struct {
		id dto.Snowflake
		Event
	}

	ch := make(chan evt, len(instances))

	for _, i := range instances {
		c := i.AttachListener(req.Msg.IncludeLogs)
		defer i.DetachListener(c)

		go func() {
			for {
				e, ok := <-c
				if !ok {
					break
				}
				ch <- evt{id: i.ID, Event: e}
			}
		}()
	}

	for {
		ev, ok := <-ch
		if !ok {
			break
		}

		err := stream.Send(&pb.RunnerListenManyResponse{
			InstanceId: uint64(ev.id),
			Event:      ev.IntoPB(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
