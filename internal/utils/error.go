package utils

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
)

func Error(c connect.Code, error string) *connect.Error {
	return connect.NewError(c, errors.New(error))
}

type unwrap interface {
	Unwrap() []error
}

type errorInterceptor struct{}

func NewErrorInterceptor() connect.Interceptor {
	return &errorInterceptor{}
}

// WrapUnary implements connect.Interceptor.
func (e *errorInterceptor) WrapUnary(handler connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		res, err := handler(ctx, req)
		if err == nil {
			return res, err
		}

		errs, ok := err.(unwrap)
		if !ok {
			return res, err
		}
		code := connect.CodeUnknown

		b := strings.Builder{}
		for i, err := range errs.Unwrap() {
			if i != 0 {
				b.WriteString(": ")
			}
			serr, ok := err.(*connect.Error)
			if ok {
				code = serr.Code()
				b.WriteString(serr.Message())
			} else {
				b.WriteString(err.Error())
			}
		}

		return res, Error(code, b.String())
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (e *errorInterceptor) WrapStreamingClient(
	handler connect.StreamingClientFunc,
) connect.StreamingClientFunc {
	return handler
}

// WrapStreamingHandler implements connect.Interceptor.
func (e *errorInterceptor) WrapStreamingHandler(
	handler connect.StreamingHandlerFunc,
) connect.StreamingHandlerFunc {
	return func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		err := handler(ctx, shc)
		if err == nil {
			return err
		}

		errs, ok := err.(unwrap)
		if !ok {
			return err
		}
		code := connect.CodeUnknown

		b := strings.Builder{}
		for i, err := range errs.Unwrap() {
			if i != 0 {
				b.WriteString(": ")
			}
			serr, ok := err.(*connect.Error)
			if ok {
				code = serr.Code()
				b.WriteString(serr.Message())
			} else {
				b.WriteString(err.Error())
			}
		}

		return Error(code, b.String())
	}
}
