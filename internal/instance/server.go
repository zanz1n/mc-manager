package instance

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ pb.InstanceServiceServer = (*Server)(nil)
var _ pb.EventServiceServer = (*Server)(nil)

type Server struct {
	m        *Manager
	versions *distribution.Repository
	pb.UnimplementedInstanceServiceServer
	pb.UnimplementedEventServiceServer
}

func NewServer(m *Manager, v *distribution.Repository) *Server {
	return &Server{m: m, versions: v}
}

// GetById implements pb.InstanceServiceServer.
func (s *Server) GetById(
	ctx context.Context,
	req *pb.InstanceGetByIdRequest,
) (*pb.Instance, error) {
	i, err := s.m.GetById(ctx, dto.Snowflake(req.Id))
	if err != nil {
		return nil, err
	}

	return i.IntoPB(), nil
}

// Launch implements pb.InstanceServiceServer.
func (s *Server) Launch(
	ctx context.Context,
	req *pb.InstanceLaunchRequest,
) (*pb.Instance, error) {
	var (
		version distribution.Version
		err     error
	)
	if req.Version == "" {
		version, err = s.versions.GetLatest(ctx, req.VersionDistro)
	} else {
		version, err = s.versions.GetVersion(ctx, req.VersionDistro, req.Version)
	}

	if err != nil {
		return nil, err
	}

	var (
		limits InstanceLimits
		config InstanceConfig
	)
	limits.FromPB(req.Limits)
	config.FromPB(req.Config)

	i, err := s.m.Launch(ctx, InstanceCreateData{
		ID:      dto.Snowflake(req.Id),
		Name:    req.Name,
		Version: version,
		Limits:  limits,
		Config:  config,
	})
	if err != nil {
		return nil, err
	}

	return i.IntoPB(), nil
}

// Stop implements pb.InstanceServiceServer.
func (s *Server) Stop(
	ctx context.Context,
	req *pb.InstanceStopRequest,
) (*pb.Instance, error) {
	i, err := s.m.GetById(ctx, dto.Snowflake(req.Id))
	if err != nil {
		return nil, err
	}

	// TODO: timeout
	err = s.m.Stop(ctx, dto.Snowflake(req.Id))
	if err != nil {
		return nil, err
	}

	return i.IntoPB(), nil
}

// Consume implements pb.EventServiceServer.
func (s *Server) Consume(
	req *pb.EventConsumeRequest,
	stream grpc.ServerStreamingServer[pb.Event],
) error {
	i, err := s.m.GetById(stream.Context(), dto.Snowflake(req.InstanceId))
	if err != nil {
		return err
	}

	ch := i.AttachListener(req.IncludeLogs)
	defer func() {
		if !i.DetachListener(ch) {
			slog.Error(
				"Instance: Failed to detach listener",
				"id", i.ID,
			)
		}
	}()

	md := metadata.MD{}
	md.Set("X-Instance-Listeners", strconv.Itoa(i.ListenersCount()))
	md.Set(
		"X-Instance-Uptime",
		time.Since(i.LaunchedAt).Round(time.Second).String(),
	)
	md.Set("X-Instance-Players", strconv.Itoa(int(i.Players.Load())))
	stream.SendHeader(md)

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

// ConsumeMany implements pb.EventServiceServer.
func (s *Server) ConsumeMany(
	req *pb.EventConsumeManyRequest,
	stream grpc.ServerStreamingServer[pb.EventConsumeManyResponse],
) error {
	instances, err := s.m.GetMany(stream.Context(), req.Instances)
	if err != nil {
		return err
	}

	type evt struct {
		id dto.Snowflake
		Event
	}

	ch := make(chan evt, len(instances))

	for _, i := range instances {
		c := i.AttachListener(req.IncludeLogs)
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

		err := stream.Send(&pb.EventConsumeManyResponse{
			InstanceId: uint64(ev.id),
			Event:      ev.IntoPB(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
