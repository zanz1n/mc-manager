package utils

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type unwrap interface {
	Unwrap() []error
}

type grpcStatus interface {
	GRPCStatus() *status.Status
}

func ErrorUnaryServerInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	res, err := handler(ctx, req)
	if err == nil {
		return res, err
	}

	errs, ok := err.(unwrap)
	if !ok {
		return res, err
	}
	code := codes.Unknown

	b := strings.Builder{}
	for i, err := range errs.Unwrap() {
		if i != 0 {
			b.WriteString(": ")
		}
		serr, ok := err.(grpcStatus)
		if ok {
			code = serr.GRPCStatus().Code()
			b.WriteString(serr.GRPCStatus().Message())
		} else {
			b.WriteString(err.Error())
		}
	}

	return res, status.Error(code, b.String())
}

func ErrorStreamServerInterceptor(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	err := handler(srv, ss)
	if err == nil {
		return err
	}

	errs, ok := err.(unwrap)
	if !ok {
		return err
	}
	code := codes.Unknown

	b := strings.Builder{}
	for i, err := range errs.Unwrap() {
		if i != 0 {
			b.WriteString(": ")
		}
		serr, ok := err.(grpcStatus)
		if ok {
			code = serr.GRPCStatus().Code()
			b.WriteString(serr.GRPCStatus().Message())
		} else {
			b.WriteString(err.Error())
		}
	}

	return status.Error(code, b.String())
}
