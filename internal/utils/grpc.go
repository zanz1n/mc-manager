package utils

import (
	"time"

	"google.golang.org/grpc"
)

func CloseGrpc(timeout time.Duration, server *grpc.Server) bool {
	ch := make(chan struct{})
	go func() {
		server.GracefulStop()
		ch <- struct{}{}
	}()

	graceful := false
	select {
	case <-ch:
		graceful = true
	case <-time.Tick(timeout):
	}

	return graceful
}
