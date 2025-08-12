package instance

import (
	"context"

	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
)

var _ pb.InstanceServiceServer = (*Server)(nil)

type Server struct {
	m        *Manager
	versions *distribution.Repository
	pb.UnsafeInstanceServiceServer
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
