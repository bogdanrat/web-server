package interceptor

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"time"
)

const (
	timeDurationFormat = "2006-01-02 15:04:05"
)

// Interceptors: execute logic (e.g., logging, auth, metrics etc...) before or after the execution of the remote function, for either client or server applications.
// To intercept a unary RPC, a function of type grpc.UnaryServerInterceptor needs to be implemented and registered to the gRPC server.

// RequestDurationInterceptor logs the request duration
func RequestDurationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Printf("%s RPC Request for %s in progress at %v\n", info.Server, info.FullMethod, start.Format(timeDurationFormat))

	m, err := handler(ctx, req)

	if err != nil {
		log.Printf("%s RPC Request for %s failed in %v\n", info.Server, info.FullMethod, time.Now().Format(timeDurationFormat))
	} else {
		log.Printf("%s RPC Request for %s fulfilled in %v\n", info.Server, info.FullMethod, time.Since(start))
	}

	return m, err
}
