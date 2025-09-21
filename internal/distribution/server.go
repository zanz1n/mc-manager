package distribution

import (
	"context"

	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/pb/pbconnect"
)

var _ pbconnect.DistributionServiceHandler = (*Server)(nil)

type Server struct {
	r *Repository

	pbconnect.UnimplementedDistributionServiceHandler
}

func NewServer(r *Repository) *Server {
	return &Server{r: r}
}

// GetLatest implements pbconnect.DistributionServiceHandler.
func (s *Server) GetLatest(
	ctx context.Context,
	req *connect.Request[pb.DistributionGetLatestRequest],
) (*connect.Response[pb.Version], error) {
	v, err := s.r.GetLatest(ctx, req.Msg.Distribution)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(v.IntoPB()), nil
}

// GetVersion implements pbconnect.DistributionServiceHandler.
func (s *Server) GetVersion(
	ctx context.Context,
	req *connect.Request[pb.DistributionGetVersionRequest],
) (*connect.Response[pb.Version], error) {
	v, err := s.r.GetVersion(ctx, req.Msg.Distribution, req.Msg.VersionId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(v.IntoPB()), nil
}

// GetAll implements pbconnect.DistributionServiceHandler.
func (s *Server) GetAll(
	ctx context.Context,
	req *connect.Request[pb.DistributionGetAllRequest],
) (*connect.Response[pb.DistributionGetAllResponse], error) {
	v, err := s.r.GetAll(ctx, req.Msg.Distribution)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.DistributionGetAllResponse{
		Versions: v,
	}), nil
}
