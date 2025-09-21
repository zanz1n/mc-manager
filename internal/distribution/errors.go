package distribution

import (
	"connectrpc.com/connect"
	"github.com/zanz1n/mc-manager/internal/utils"
)

var (
	ErrHttp = utils.Error(
		connect.CodeInternal,
		"http error while fetching distribution",
	)
	ErrVersionNotFound = utils.Error(
		connect.CodeNotFound,
		"distribution version not found",
	)
	ErrInvalidDistribution = utils.Error(
		connect.CodeInvalidArgument,
		"distribution is invalid",
	)

	ErrHashNotAvailable = utils.Error(
		connect.CodeUnavailable,
		"hash not available for this version",
	)
	ErrHashFailed = utils.Error(
		connect.CodeInternal,
		"failed to verify hash",
	)
)
