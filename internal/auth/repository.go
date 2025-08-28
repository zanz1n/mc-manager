package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Respository struct {
	token      string
	a          Auther
	expiration time.Duration

	db db.Querier
}

func NewRepository(a Auther, db db.Querier, cfg *config.APIConfig) *Respository {
	return &Respository{
		token:      cfg.Server.Password,
		a:          a,
		expiration: cfg.Auth.JWTExpiration,
		db:         db,
	}
}

var _ Authed = authedServer{}

type authedServer struct{}

// GetId implements Authed.
func (a authedServer) GetId() dto.Snowflake {
	return dto.NullSnowflake
}

// IsAdmin implements Authed.
func (a authedServer) IsAdmin() bool {
	return true
}

func (r *Respository) Authenticate(ctx context.Context) (Authed, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || md == nil {
		return nil, ErrInvalidAuthToken
	}

	ttype, token, err := getMeta(md)
	if err != nil {
		return nil, err
	}

	switch ttype {
	case "Bearer":
		authed, err := r.authUser(ctx, token, md)
		if err != nil {
			return nil, err
		}
		return &authed, nil

	case "Server", "SRV":
		return r.authServer(token)

	default:
		return nil, errors.Join(
			ErrInvalidAuthToken,
			fmt.Errorf("unknown auth strategy `%s`", ttype),
		)
	}
}

func (r *Respository) AuthenticateUser(ctx context.Context) (Token, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || md == nil {
		return Token{}, ErrInvalidAuthToken
	}

	ttype, token, err := getMeta(md)
	if err != nil {
		return Token{}, err
	}

	switch ttype {
	case "Bearer":
		return r.authUser(ctx, token, md)

	case "Server", "SRV":
		return Token{}, errors.Join(
			ErrInvalidAuthToken,
			fmt.Errorf("auth strategy `%s` not valid for this method", ttype),
		)

	default:
		return Token{}, errors.Join(
			ErrInvalidAuthToken,
			fmt.Errorf("unknown auth strategy `%s`", ttype),
		)
	}
}

func (r *Respository) authServer(
	tokenstr string,
) (Authed, error) {
	if r.token == "" {
		return nil, errors.Join(
			ErrInvalidAuthToken,
			fmt.Errorf("server auth strategy disabled"),
		)
	}

	if r.token != tokenstr {
		return nil, errors.Join(
			ErrInvalidAuthToken,
			fmt.Errorf("server token mismatches"),
		)
	}

	return authedServer{}, nil
}

func (r *Respository) authUser(
	ctx context.Context,
	tokenstr string,
	md metadata.MD,
) (token Token, err error) {
	token, err = r.a.DecodeToken(tokenstr)
	if err == nil {
		return
	}

	rthead := md.Get("auth-refresh-token")
	if len(rthead) != 1 {
		err = ErrInvalidAuthToken
		return
	}

	userId, err := r.a.ValidateRefreshToken(ctx, rthead[0])
	if err != nil {
		return
	}

	user, err := r.db.UserGetById(ctx, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrUserNotFound
		}
		return
	}
	token = NewToken(user, "", r.expiration)

	tokenstr, err = r.a.EncodeToken(token)
	if err != nil {
		return
	}

	err = grpc.SendHeader(ctx, metadata.MD{
		"set-token": []string{tokenstr},
	})
	return
}

func getMeta(md metadata.MD) (ttype string, t string, err error) {
	h := md.Get("authorization")
	if len(h) != 1 {
		return "", "", ErrInvalidAuthToken
	}

	var ok bool
	ttype, t, ok = strings.Cut(h[0], " ")
	if !ok || md == nil {
		return "", "", ErrInvalidAuthToken
	}
	return
}
