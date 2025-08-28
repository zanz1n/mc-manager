package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/dto"
)

type Auther interface {
	EncodeToken(data Token) (string, error)
	DecodeToken(token string) (Token, error)

	ValidateRefreshToken(ctx context.Context, token string) (dto.Snowflake, error)
	GenRefreshToken(ctx context.Context, userId dto.Snowflake) (string, error)
	DeleteRefreshTokens(ctx context.Context, userId dto.Snowflake) error
}

type Authed interface {
	// can return empty
	GetId() dto.Snowflake
	IsAdmin() bool
}

var _ Authed = (*Token)(nil)
var _ jwt.Claims = (*Token)(nil)

type Token struct {
	ID        dto.Snowflake   `json:"sub"`
	IssuedAt  jwt.NumericDate `json:"iat"`
	ExpiresAt jwt.NumericDate `json:"exp"`
	Issuer    string          `json:"iss"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	Admin     bool            `json:"admin"`
}

func NewToken(user db.User, issuer string, exp time.Duration) Token {
	now := time.Now().Round(time.Second)

	return Token{
		ID:        user.ID,
		IssuedAt:  jwt.NumericDate{Time: now},
		ExpiresAt: jwt.NumericDate{Time: now.Add(exp)},
		Issuer:    issuer,
		Username:  user.Username,
		Email:     user.Email,
		Admin:     user.Admin,
	}
}

// GetId implements Authed.
func (t *Token) GetId() dto.Snowflake {
	return t.ID
}

// IsAdmin implements Authed.
func (t *Token) IsAdmin() bool {
	return t.Admin
}

// GetAudience implements jwt.Claims.
func (t *Token) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// GetExpirationTime implements jwt.Claims.
func (t *Token) GetExpirationTime() (*jwt.NumericDate, error) {
	return &t.ExpiresAt, nil
}

// GetIssuedAt implements jwt.Claims.
func (t *Token) GetIssuedAt() (*jwt.NumericDate, error) {
	return &t.IssuedAt, nil
}

// GetIssuer implements jwt.Claims.
func (t *Token) GetIssuer() (string, error) {
	return t.Issuer, nil
}

// GetNotBefore implements jwt.Claims.
func (t *Token) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetSubject implements jwt.Claims.
func (t *Token) GetSubject() (string, error) {
	return t.ID.String(), nil
}
