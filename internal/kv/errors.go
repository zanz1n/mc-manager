package kv

import "errors"

var (
	ErrValueNotFound   = errors.New("key value: value not found")
	ErrValueNotPointer = errors.New("key value: v must be a pointer")
)
