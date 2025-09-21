package auth

import (
	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/internal/utils"
)

var (
	ErrExpiredAuthToken = utils.Error(
		connect.CodeUnauthenticated,
		"authentication token expired",
	)

	ErrInvalidAuthToken = utils.Error(
		connect.CodeUnauthenticated,
		"authentication token is invalid or was not provided",
	)

	ErrInvalidRefreshToken = utils.Error(
		connect.CodeUnauthenticated,
		"refresh token invalid",
	)

	ErrUserNotFound = utils.Error(
		connect.CodeNotFound,
		"user not found",
	)
)
