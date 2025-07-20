package distribution

import "errors"

var (
	ErrHttp            = errors.New("http error while fetching distribution")
	ErrVersionNotFound = errors.New("distribution version not found")

	ErrHashNotAvailable = errors.New("hash not available for this version")
	ErrHashFailed       = errors.New("failed to verify hash")
)
