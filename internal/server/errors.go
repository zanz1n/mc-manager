package server

import (
	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/internal/utils"
)

var (
	ErrInstanceNotFound = utils.Error(
		connect.CodeNotFound,
		"instance not found",
	)

	ErrNodeNotFound = utils.Error(
		connect.CodeNotFound,
		"node not found",
	)

	ErrNodeUnreachable = utils.Error(
		connect.CodeInternal,
		"node is unreachable",
	)

	ErrLocalNodeUndeletable = utils.Error(
		connect.CodePermissionDenied,
		"local node can not be deleted",
	)

	ErrPermissionDenied = utils.Error(
		connect.CodePermissionDenied,
		"permission denied",
	)

	ErrUserNotFound = utils.Error(
		connect.CodeNotFound,
		"user not found",
	)

	ErrLogin = utils.Error(
		connect.CodePermissionDenied,
		"user does not exist or password mismatches",
	)
)
