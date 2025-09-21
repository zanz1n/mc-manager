package runner

import (
	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/internal/utils"
)

var (
	ErrJavaVersion = utils.Error(
		connect.CodeNotFound,
		"the instance java version is invalid",
	)
	ErrFileSystem = utils.Error(
		connect.CodeInternal,
		"filesystem error",
	)
	ErrInstanceNotFound = utils.Error(
		connect.CodeNotFound,
		"instance not found",
	)
	ErrInstanceAlreadyLaunched = utils.Error(
		connect.CodeAlreadyExists,
		"instance already launched",
	)
	ErrInstanceCreate = utils.Error(
		connect.CodeInternal,
		"failed to create instance",
	)
	ErrInstanceLaunch = utils.Error(
		connect.CodeInternal,
		"failed to launch instance",
	)
	ErrInstanceStop = utils.Error(
		connect.CodeInternal,
		"failed to stop instance",
	)
	ErrInvalidCreateData = utils.Error(
		connect.CodeInvalidArgument,
		"invalid instance create data",
	)
	ErrSendCommand = utils.Error(
		connect.CodeInternal,
		"failed to send command to instance",
	)
)
