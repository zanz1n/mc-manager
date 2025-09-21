package server

import (
	"context"
	"database/sql"
	"errors"

	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/auth"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/pb/pbconnect"
	"golang.org/x/crypto/bcrypt"
)

var _ pbconnect.UserServiceHandler = (*UserServer)(nil)

type UserServer struct {
	db         db.Querier
	ar         *auth.Respository
	bcryptCost int

	// pb.UnimplementedUserServiceServer
}

func NewUserServer(
	db db.Querier,
	ar *auth.Respository,
	cfg *config.APIConfig,
) *UserServer {
	return &UserServer{
		db:         db,
		ar:         ar,
		bcryptCost: int(cfg.Auth.BcryptCost),
	}
}

// GetById implements pbconnect.UserServiceHandler.
func (s *UserServer) GetById(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[pb.User], error) {
	res := connect.NewResponse((*pb.User)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	id := dto.Snowflake(req.Msg.Id)

	if !authed.IsAdmin() {
		if id != authed.GetId() {
			return nil, ErrPermissionDenied
		}
	}

	user, err := s.db.UserGetById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrUserNotFound, errors.New(id.String()))
		}
		return nil, err
	}

	res.Msg = user.IntoPB()
	return res, nil
}

// GetMany implements pbconnect.UserServiceHandler.
func (s *UserServer) GetMany(
	ctx context.Context,
	req *connect.Request[pb.Pagination],
) (*connect.Response[pb.UserGetManyResponse], error) {
	res := connect.NewResponse((*pb.UserGetManyResponse)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	lastSeen := dto.Snowflake(req.Msg.LastSeen)
	users, err := s.db.UserGetMany(ctx, lastSeen, req.Msg.Limit)
	if err != nil {
		return nil, err
	}

	pbusers := make([]*pb.User, len(users))
	for i, u := range users {
		pbusers[i] = u.IntoPB()
	}

	res.Msg = &pb.UserGetManyResponse{Users: pbusers}
	return res, nil
}

// Create implements pbconnect.UserServiceHandler.
func (s *UserServer) Create(
	ctx context.Context,
	req *connect.Request[pb.UserCreateRequest],
) (*connect.Response[pb.User], error) {
	res := connect.NewResponse((*pb.User)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(req.Msg.Password),
		s.bcryptCost,
	)
	if err != nil {
		return nil, err
	}

	user, err := s.db.UserCreate(ctx, db.UserCreateParams{
		ID:            dto.NewSnowflake(),
		Username:      req.Msg.Username,
		FirstName:     req.Msg.FirstName,
		LastName:      req.Msg.LastName,
		MinecraftUser: req.Msg.MinecraftUser,
		Email:         req.Msg.Email,
		Admin:         req.Msg.Admin,
		TwoFa:         req.Msg.TwoFa,
		Password:      hashed,
	})
	if err != nil {
		return nil, err
	}

	res.Msg = user.IntoPB()
	return res, nil
}

// Delete implements pbconnect.UserServiceHandler.
func (s *UserServer) Delete(
	ctx context.Context,
	req *connect.Request[pb.Snowflake],
) (*connect.Response[pb.User], error) {
	res := connect.NewResponse((*pb.User)(nil))

	authed, err := s.ar.Authenticate(ctx, req.Header(), res.Header())
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	id := dto.Snowflake(req.Msg.Id)
	user, err := s.db.UserDelete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrUserNotFound, errors.New(id.String()))
		}
		return nil, err
	}

	res.Msg = user.IntoPB()
	return res, nil
}
