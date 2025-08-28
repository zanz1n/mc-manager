package server

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInstanceNotFound = status.Error(
		codes.NotFound,
		"instance not found",
	)

	ErrNodeNotFound = status.Error(
		codes.NotFound,
		"node not found",
	)

	ErrNodeUnreachable = status.Error(
		codes.Internal,
		"node is unreachable",
	)

	ErrPermissionDenied = status.Error(
		codes.PermissionDenied,
		"permission denied",
	)

	ErrUserNotFound = status.Error(
		codes.NotFound,
		"user not found",
	)

	ErrLogin = status.Error(
		codes.PermissionDenied,
		"user does not exist or password mismatches",
	)
)
