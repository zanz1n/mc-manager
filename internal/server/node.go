package server

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zanz1n/mc-manager/internal/auth"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ pb.NodeServiceServer = (*NodeServer)(nil)

type NodeServer struct {
	db          db.Querier
	ar          *auth.Respository
	localNodeId dto.Snowflake

	pb.UnimplementedNodeServiceServer
}

func NewNodeServer(db db.Querier, ar *auth.Respository, localNode dto.Snowflake) *NodeServer {
	return &NodeServer{
		db:          db,
		ar:          ar,
		localNodeId: localNode,
	}
}

// GetById implements pb.NodeServiceServer.
func (s *NodeServer) GetById(ctx context.Context, req *pb.Snowflake) (*pb.Node, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	if req.Id == uint64(s.localNodeId) {
		return s.localNodeInformation(), nil
	}

	id := dto.Snowflake(req.Id)
	node, err := s.db.NodeGetById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrNodeNotFound, errors.New(id.String()))
		}
		return nil, err
	}

	return node.IntoPB(), nil
}

// GetMany implements pb.NodeServiceServer.
func (s *NodeServer) GetMany(
	ctx context.Context,
	req *pb.Pagination,
) (*pb.NodeGetManyResponse, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	nodes, err := s.db.NodeGetMany(ctx, dto.Snowflake(req.LastSeen), req.Limit)
	if err != nil {
		return nil, err
	}

	res := make([]*pb.Node, len(nodes))
	for i, node := range nodes {
		res[i] = node.IntoPB()
	}

	return &pb.NodeGetManyResponse{
		Nodes: res,
	}, nil
}

// Create implements pb.NodeServiceServer.
func (s *NodeServer) Create(ctx context.Context, req *pb.NodeCreateRequest) (*pb.Node, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}
	id := dto.NewSnowflake()

	node, err := s.db.NodeCreate(ctx, db.NodeCreateParams{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Token:       req.Token,
		Endpoint:    req.Endpoint,
		EndpointTls: req.EndpointTls,
		FtpPort:     int32(req.FtpPort),
		GrpcPort:    int32(req.GrpcPort),
	})
	if err != nil {
		return nil, err
	}

	return node.IntoPB(), nil
}

// Delete implements pb.NodeServiceServer.
func (s *NodeServer) Delete(ctx context.Context, req *pb.Snowflake) (*pb.Node, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}
	id := dto.Snowflake(req.Id)

	if id == s.localNodeId {
		return nil, ErrLocalNodeUndeletable
	}

	node, err := s.db.NodeDelete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrNodeNotFound, errors.New(id.String()))
		}
		return nil, err
	}

	return node.IntoPB(), nil
}

func (s *NodeServer) localNodeInformation() *pb.Node {
	now := time.Now()

	return &pb.Node{
		Id:          uint64(s.localNodeId),
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
		Name:        "Local Node",
		Description: "",
		Maintenance: false,
		Token:       "",
		Endpoint:    "localhost",
		EndpointTls: false,
		FtpPort:     0,
		GrpcPort:    0,
	}
}
