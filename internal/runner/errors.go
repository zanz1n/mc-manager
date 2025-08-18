package runner

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrJavaVersion = status.Error(
		codes.NotFound,
		"the instance java version is invalid",
	)
	ErrFileSystem = status.Error(
		codes.Internal,
		"filesystem error",
	)
	ErrInstanceNotFound = status.Error(
		codes.NotFound,
		"instance not found",
	)
	ErrInstanceAlreadyLaunched = status.Error(
		codes.AlreadyExists,
		"instance already launched",
	)
	ErrInstanceCreate = status.Error(
		codes.Internal,
		"failed to create instance",
	)
	ErrInstanceLaunch = status.Error(
		codes.Internal,
		"failed to launch instance",
	)
	ErrInstanceStop = status.Error(
		codes.Internal,
		"failed to stop instance",
	)
	ErrInvalidCreateData = status.Error(
		codes.InvalidArgument,
		"invalid instance create data",
	)
	ErrSendCommand = status.Error(
		codes.Internal,
		"failed to send command to instance",
	)
)
