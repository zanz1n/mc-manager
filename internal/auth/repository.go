package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
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

func (r *Respository) Authenticate(
	ctx context.Context,
	reqHead, resHead http.Header,
) (Authed, error) {
	ttype, token, err := getMeta(reqHead)
	if err != nil {
		return nil, err
	}

	switch ttype {
	case "Bearer":
		authed, err := r.authUser(ctx, token, reqHead, resHead)
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

func (r *Respository) AuthenticateUser(
	ctx context.Context,
	reqHead, resHead http.Header,
) (Token, error) {
	ttype, token, err := getMeta(reqHead)
	if err != nil {
		return Token{}, err
	}

	switch ttype {
	case "Bearer":
		return r.authUser(ctx, token, reqHead, resHead)

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
	reqHead, resHead http.Header,
) (token Token, err error) {
	token, err = r.a.DecodeToken(tokenstr)
	if err == nil {
		return
	}

	rthead := reqHead.Values("auth-refresh-token")
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

	resHead.Set("set-token", tokenstr)

	return
}

func getMeta(head http.Header) (ttype string, t string, err error) {
	h := head.Values("authorization")
	if len(h) != 1 {
		return "", "", ErrInvalidAuthToken
	}

	var ok bool
	ttype, t, ok = strings.Cut(h[0], " ")
	if !ok || head == nil {
		return "", "", ErrInvalidAuthToken
	}
	return
}
