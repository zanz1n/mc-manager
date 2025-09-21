package utils

import (
	"context"

	"connectrpc.com/connect"
)

var (
	errMissingAuthorization = Error(
		connect.CodeUnauthenticated,
		"no incoming `authorization` metadata in grpc context",
	)
	errPasswordMismatches = Error(
		connect.CodeUnauthenticated,
		"the `authorization` metadata password mismatches",
	)
)

type authInterceptor struct {
	passwd string
}

func NewAuthInterceptor(passwd string) connect.Interceptor {
	return &authInterceptor{passwd: passwd}
}

// WrapUnary implements connect.Interceptor.
func (a *authInterceptor) WrapUnary(handler connect.UnaryFunc) connect.UnaryFunc {
	if a.passwd == "" {
		return handler
	}

	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		authHead := req.Header().Get("authorization")
		if authHead == "" {
			return nil, errMissingAuthorization
		}

		if authHead != a.passwd {
			return nil, errPasswordMismatches
		}
		return handler(ctx, req)
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (a *authInterceptor) WrapStreamingClient(
	handler connect.StreamingClientFunc,
) connect.StreamingClientFunc {
	if a.passwd == "" {
		return handler
	}

	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := handler(ctx, spec)
		conn.RequestHeader().Add("authorization", a.passwd)
		return conn
	}
}

// WrapStreamingHandler implements connect.Interceptor.
func (a *authInterceptor) WrapStreamingHandler(
	handler connect.StreamingHandlerFunc,
) connect.StreamingHandlerFunc {
	if a.passwd == "" {
		return handler
	}

	return func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		authHead := shc.RequestHeader().Get("authorization")
		if authHead == "" {
			return errMissingAuthorization
		}

		if authHead != a.passwd {
			return errPasswordMismatches
		}
		return handler(ctx, shc)
	}
}
