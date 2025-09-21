package server

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/auth"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
	"github.com/zanz1n/mc-manager/internal/pb/pbconnect"
	"github.com/zanz1n/mc-manager/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ pbconnect.AuthServiceHandler = (*AuthServer)(nil)

type AuthServer struct {
	db           db.Querier
	a            auth.Auther
	ar           *auth.Respository
	expiration   time.Duration
	enableSignup bool
	bcryptCost   int

	// pbconnect.UnimplementedAuthServiceHandler
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

// GetSelf implements pbconnect.AuthServiceHandler.
func (s *AuthServer) GetSelf(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[pb.User], error) {
	res := connect.NewResponse((*pb.User)(nil))

	token, err := s.ar.AuthenticateUser(ctx, req.Header(), res.Header())
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

	res.Msg = user.IntoPB()
	return res, nil
}

// Login implements pbconnect.AuthServiceHandler.
func (s *AuthServer) Login(
	ctx context.Context,
	req *connect.Request[pb.AuthLoginRequest],
) (*connect.Response[pb.AuthLoginResponse], error) {
	user, err := s.db.UserGetByEmail(ctx, req.Msg.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrLogin
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(req.Msg.Password))
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

	return connect.NewResponse(&pb.AuthLoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}), nil
}

// Signup implements pbconnect.AuthServiceHandler.
func (s *AuthServer) Signup(
	ctx context.Context,
	req *connect.Request[pb.AuthSignupRequest],
) (*connect.Response[pb.AuthSignupResponse], error) {
	if !s.enableSignup {
		return nil, utils.Error(
			connect.CodePermissionDenied,
			"signup is disabled",
		)
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

	return connect.NewResponse(&pb.AuthSignupResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user.IntoPB(),
	}), nil
}
