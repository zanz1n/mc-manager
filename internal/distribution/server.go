package distribution

import (
	"context"

	"github.com/zanz1n/mc-manager/internal/pb"
)

var _ pb.DistributionServiceServer = (*Server)(nil)

type Server struct {
	r *Repository
	pb.UnimplementedDistributionServiceServer
}

func NewServer(r *Repository) *Server {
	return &Server{r: r}
}

// GetLatest implements pb.DistributionServiceServer.
func (s *Server) GetLatest(
	ctx context.Context,
	req *pb.DistributionGetLatestRequest,
) (*pb.Version, error) {
	v, err := s.r.GetLatest(ctx, req.Distribution)
	if err != nil {
		return nil, err
	}

	return v.IntoPB(), nil
}

// GetVersion implements pb.DistributionServiceServer.
func (s *Server) GetVersion(
	ctx context.Context,
	req *pb.DistributionGetVersionRequest,
) (*pb.Version, error) {
	v, err := s.r.GetVersion(ctx, req.Distribution, req.VersionId)
	if err != nil {
		return nil, err
	}

	return v.IntoPB(), nil
}

// GetAll implements pb.DistributionServiceServer.
func (s *Server) GetAll(
	ctx context.Context,
	req *pb.DistributionGetAllRequest,
) (*pb.DistributionGetAllResponse, error) {
	v, err := s.r.GetAll(ctx, req.Distribution)
	if err != nil {
		return nil, err
	}

	return &pb.DistributionGetAllResponse{
		Versions: v,
	}, nil
}
