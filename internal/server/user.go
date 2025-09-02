package server

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/auth"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"golang.org/x/crypto/bcrypt"
)

var _ pb.UserServiceServer = (*UserServer)(nil)

type UserServer struct {
	db         db.Querier
	ar         *auth.Respository
	bcryptCost int

	pb.UnimplementedUserServiceServer
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

// GetById implements pb.UserServiceServer.
func (s *UserServer) GetById(ctx context.Context, req *pb.Snowflake) (*pb.User, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	id := dto.Snowflake(req.Id)

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
	return user.IntoPB(), nil
}

// GetMany implements pb.UserServiceServer.
func (s *UserServer) GetMany(
	ctx context.Context,
	req *pb.Pagination,
) (*pb.UserGetManyResponse, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	lastSeen := dto.Snowflake(req.LastSeen)
	users, err := s.db.UserGetMany(ctx, lastSeen, req.Limit)
	if err != nil {
		return nil, err
	}

	pbusers := make([]*pb.User, len(users))
	for i, u := range users {
		pbusers[i] = u.IntoPB()
	}

	return &pb.UserGetManyResponse{
		Users: pbusers,
	}, nil
}

// Create implements pb.UserServiceServer.
func (s *UserServer) Create(ctx context.Context, req *pb.UserCreateRequest) (*pb.User, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		s.bcryptCost,
	)
	if err != nil {
		return nil, err
	}

	user, err := s.db.UserCreate(ctx, db.UserCreateParams{
		ID:            dto.NewSnowflake(),
		Username:      req.Username,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		MinecraftUser: req.MinecraftUser,
		Email:         req.Email,
		Admin:         req.Admin,
		TwoFa:         req.TwoFa,
		Password:      hashed,
	})
	if err != nil {
		return nil, err
	}
	return user.IntoPB(), nil
}

// Delete implements pb.UserServiceServer.
func (s *UserServer) Delete(ctx context.Context, req *pb.Snowflake) (*pb.User, error) {
	authed, err := s.ar.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if !authed.IsAdmin() {
		return nil, ErrPermissionDenied
	}

	id := dto.Snowflake(req.Id)
	user, err := s.db.UserDelete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Join(ErrUserNotFound, errors.New(id.String()))
		}
		return nil, err
	}

	return user.IntoPB(), nil
}
