package auth

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrExpiredAuthToken = status.Error(
		codes.Unauthenticated,
		"authentication token expired",
	)

	ErrInvalidAuthToken = status.Error(
		codes.Unauthenticated,
		"authentication token is invalid or was not provided",
	)

	ErrInvalidRefreshToken = status.Error(
		codes.Unauthenticated,
		"refresh token invalid",
	)

	ErrUserNotFound = status.Error(
		codes.NotFound,
		"user not found",
	)
)
