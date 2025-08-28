package auth

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/kv"
)

var _ Auther = (*JWTAuther)(nil)

type JWTAuther struct {
	issuer     string
	parser     *jwt.Parser
	expiration time.Duration

	kv kv.KVStorer

	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

func NewJWTAuther(
	kv kv.KVStorer,
	priv ed25519.PrivateKey,
	pub ed25519.PublicKey,
	cfg *config.APIConfig,
) *JWTAuther {
	iss := cfg.Name
	if cfg.Name == "" {
		iss = "SRV"
	}

	return &JWTAuther{
		issuer:     iss,
		parser:     jwt.NewParser(),
		expiration: cfg.Auth.JWTExpiration,
		kv:         kv,
		priv:       priv,
		pub:        pub,
	}
}

// EncodeToken implements Auther.
func (a *JWTAuther) EncodeToken(data Token) (string, error) {
	if data.Issuer == "" {
		data.Issuer = a.issuer
	}

	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &data)
	return t.SignedString(a.priv)
}

// DecodeToken implements Auther.
func (a *JWTAuther) DecodeToken(token string) (Token, error) {
	var claims Token

	t, err := a.parser.ParseWithClaims(token, &claims, a.keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return claims, ErrExpiredAuthToken
		} else {
			return claims, ErrInvalidAuthToken
		}
	}

	if !t.Valid {
		return claims, ErrInvalidAuthToken
	}

	return claims, nil
}

// ValidateRefreshToken implements Auther.
func (a *JWTAuther) ValidateRefreshToken(
	ctx context.Context,
	token string,
) (dto.Snowflake, error) {
	tokenb, err := decodeRefreshToken(token)
	if err != nil {
		return 0, err
	}

	userId := getRefreshTokenUser(tokenb)
	key := fmt.Sprintf("refresh_token/%s", userId)

	refreshToken, err := a.kv.GetEx(ctx, key, a.expiration)
	if err != nil {
		if errors.Is(err, kv.ErrValueNotFound) {
			err = ErrInvalidRefreshToken
		}
		return 0, err
	}

	if refreshToken != token {
		return 0, ErrInvalidRefreshToken
	}

	return userId, nil
}

// GenRefreshToken implements Auther.
func (a *JWTAuther) GenRefreshToken(
	ctx context.Context,
	userId dto.Snowflake,
) (string, error) {
	key := fmt.Sprintf("refresh_token/%s", userId)

	refreshToken, err := a.kv.GetEx(ctx, key, a.expiration)
	if err == nil {
		return refreshToken, nil
	}

	if !errors.Is(err, kv.ErrValueNotFound) {
		return "", err
	}

	refreshToken = generateRefreshToken(userId)

	err = a.kv.SetEx(ctx, key, refreshToken, a.expiration)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

// DeleteRefreshTokens implements Auther.
func (a *JWTAuther) DeleteRefreshTokens(
	ctx context.Context,
	userId dto.Snowflake,
) error {
	key := fmt.Sprintf("refresh_token/%s", userId)

	err := a.kv.Delete(ctx, key)
	if errors.Is(err, kv.ErrValueNotFound) {
		err = nil
	}

	return err
}

func (a *JWTAuther) keyFunc(t *jwt.Token) (any, error) {
	if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
		return nil, jwt.ErrEd25519Verification
	}

	return a.pub, nil
}
