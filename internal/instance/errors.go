package instance

import "errors"

var (
	ErrJavaVersion       = errors.New("the instance java version is invalid")
	ErrFileSystem        = errors.New("filesystem error")
	ErrInstanceNotFound  = errors.New("instance not found")
	ErrInstanceCreate    = errors.New("failed to create instance")
	ErrInstanceLaunch    = errors.New("failed to launch instance")
	ErrInstanceStop      = errors.New("failed to stop instance")
	ErrInvalidCreateData = errors.New("invalid instance create data")
	ErrSendCommand       = errors.New("failed to send command to instance")
)
