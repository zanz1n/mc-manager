package server

import (
	"context"
	"database/sql"
	"errors"

	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/internal/auth"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/pb/pbconnect"
)

var _ pbconnect.NodeServiceHandler = (*NodeServer)(nil)

type NodeServer struct {
	db          db.Querier
	ar          *auth.Respository
	localNodeId dto.Snowflake

	pbconnect.UnimplementedNodeServiceHandler
}

func NewNodeServer(db db.Querier, ar *auth.Respository, localNode dto.Snowflake) *NodeServer {
	return &NodeServer{
		db:          db,
		ar:          ar,
		localNodeId: localNode,
	}
}

// GetById implements pbconnect.NodeServiceHandler.
func (s *NodeServer) GetById(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[pb.Node], error) {
	res := connect.NewResponse((*pb.Node)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}
	id := dto.Snowflake(req.Msg.Id)

	node, err := s.db.NodeGetById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrNodeNotFound, errors.New(id.String()))
		}
		return nil, err
	}

	res.Msg = node.IntoPB()
	return res, nil
}

// GetMany implements pbconnect.NodeServiceHandler.
func (s *NodeServer) GetMany(
	ctx context.Context,
	req *connect.Request[pb.Pagination],
) (*connect.Response[pb.NodeGetManyResponse], error) {
	res := connect.NewResponse((*pb.NodeGetManyResponse)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	nodes, err := s.db.NodeGetMany(ctx, dto.Snowflake(req.Msg.LastSeen), req.Msg.Limit)
	if err != nil {
		return nil, err
	}

	resnodes := make([]*pb.Node, len(nodes))
	for i, node := range nodes {
		resnodes[i] = node.IntoPB()
	}

	res.Msg = &pb.NodeGetManyResponse{Nodes: resnodes}
	return res, nil
}

// Create implements pbconnect.NodeServiceHandler.
func (s *NodeServer) Create(
	ctx context.Context,
	req *connect.Request[pb.NodeCreateRequest],
) (*connect.Response[pb.Node], error) {
	res := connect.NewResponse((*pb.Node)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}
	id := dto.NewSnowflake()

	node, err := s.db.NodeCreate(ctx, db.NodeCreateParams{
		ID:          id,
		Name:        req.Msg.Name,
		Description: req.Msg.Description,
		Token:       req.Msg.Token,
		Endpoint:    req.Msg.Endpoint,
		EndpointTls: req.Msg.EndpointTls,
		FtpPort:     int32(req.Msg.FtpPort),
		GrpcPort:    int32(req.Msg.GrpcPort),
	})
	if err != nil {
		return nil, err
	}

	res.Msg = node.IntoPB()
	return res, nil
}

// Delete implements pbconnect.NodeServiceHandler.
func (s *NodeServer) Delete(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[pb.Node], error) {
	res := connect.NewResponse((*pb.Node)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}
	id := dto.Snowflake(req.Msg.Id)

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

	res.Msg = node.IntoPB()
	return res, nil
}
