package utils

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

type loggerInterceptor struct{}

func NewLoggerInterceptor() connect.Interceptor {
	return &loggerInterceptor{}
}

// WrapUnary implements connect.Interceptor.
func (l *loggerInterceptor) WrapUnary(handler connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()

		res, err := handler(ctx, req)
		took := time.Since(start).Round(time.Microsecond)

		if err != nil {
			s, ok := err.(*connect.Error)

			if ok {
				slog.Info(
					"GRPC: Handled unary call",
					"method", req.HTTPMethod(),
					"code", s.Code(),
					"took", took,
				)
			} else {
				slog.Info(
					"GRPC: Handled unary call with error",
					"method", req.HTTPMethod(),
					"took", took,
					"error", err,
				)
			}

		} else {
			slog.Info(
				"GRPC: Handled unary call",
				"method", req.HTTPMethod(),
				"code", "ok",
				"took", took,
			)
		}

		return res, err
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (l *loggerInterceptor) WrapStreamingClient(
	handler connect.StreamingClientFunc,
) connect.StreamingClientFunc {
	return handler
}

// WrapStreamingHandler implements connect.Interceptor.
func (l *loggerInterceptor) WrapStreamingHandler(
	handler connect.StreamingHandlerFunc,
) connect.StreamingHandlerFunc {
	return func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		start := time.Now()

		err := handler(ctx, shc)
		took := time.Since(start).Round(time.Microsecond)

		spec := shc.Spec()

		if err != nil {
			s, ok := err.(*connect.Error)

			if ok {
				slog.Info(
					"GRPC: Handled stream call",
					"method", spec.Procedure,
					"stream_type", spec.StreamType,
					"code", s.Code(),
					"took", took,
				)
			} else {
				slog.Info(
					"GRPC: Handled stream call with error",
					"method", spec.Procedure,
					"stream_type", spec.StreamType,
					"took", took,
					"error", err,
				)
			}
		} else {
			slog.Info(
				"GRPC: Handled stream call",
				"method", spec.Procedure,
				"stream_type", spec.StreamType,
				"code", "ok",
				"took", took,
			)
		}

		return err
	}
}
