package server

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/auth"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ pb.AuthServiceServer = (*AuthServer)(nil)

type AuthServer struct {
	db           db.Querier
	a            auth.Auther
	ar           *auth.Respository
	expiration   time.Duration
	enableSignup bool
	bcryptCost   int

	pb.UnimplementedAuthServiceServer
}

func NewAuthServer(
	db db.Querier,
	a auth.Auther,
	ar *auth.Respository,
	cfg *config.APIConfig,
) *AuthServer {
	return &AuthServer{
		db:           db,
		a:            a,
		ar:           ar,
		expiration:   cfg.Auth.JWTExpiration,
		enableSignup: cfg.Auth.AllowSignup,
		bcryptCost:   int(cfg.Auth.BcryptCost),
	}
}

// GetSelf implements pb.AuthServiceServer.
func (s *AuthServer) GetSelf(ctx context.Context, req *emptypb.Empty) (*pb.User, error) {
	token, err := s.ar.AuthenticateUser(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.db.UserGetById(ctx, token.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrUserNotFound
		}
		return nil, err
	}

	return user.IntoPB(), nil
}

// Login implements pb.AuthServiceServer.
func (s *AuthServer) Login(
	ctx context.Context,
	req *pb.AuthLoginRequest,
) (*pb.AuthLoginResponse, error) {
	user, err := s.db.UserGetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrLogin
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(req.Password))
	if err != nil {
		return nil, ErrLogin
	}

	token, err := s.a.EncodeToken(auth.NewToken(user, "", s.expiration))
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.a.GenRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &pb.AuthLoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// Signup implements pb.AuthServiceServer.
func (s *AuthServer) Signup(
	ctx context.Context,
	req *pb.AuthSignupRequest,
) (*pb.AuthSignupResponse, error) {
	if !s.enableSignup {
		return nil, status.Error(
			codes.PermissionDenied,
			"signup is disabled",
		)
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
		Admin:         false,
		TwoFa:         false,
		Password:      hashed,
	})
	if err != nil {
		return nil, err
	}

	token, err := s.a.EncodeToken(auth.NewToken(user, "", s.expiration))
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.a.GenRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &pb.AuthSignupResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user.IntoPB(),
	}, nil
}
