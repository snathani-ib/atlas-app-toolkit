package requestid

import (
	"context"
	"errors"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Extract Ophid from Context
func extractHeaderFromMetaData(ctx context.Context, field string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("getting grpc metadata")
	}
	if header, ok2 := md[field]; ok2 && len(header) > 0 {
		return header[0], nil
	}
	return "", errors.New("header " + field + " not in metadata.")
}

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware to generate/include Request-Id in headers and context
// for tracing and tracking user's request.
//
// Returned middleware populates Request-Id from gRPC metadata if
// they defined in a testRequest message else creates a new one.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {

		reqID := HandleRequestID(ctx)

		ophid, _ := extractHeaderFromMetaData(ctx, "ophid")
		// add request id to logger
		addRequestIDToLogger(ctx, reqID)
		addOphIDToLogger(ctx, ophid)

		ctx = NewContext(ctx, reqID)
		ctx = NewContextWithOphID(ctx, ophid)

		// returning from the request call
		res, err = handler(ctx, req)

		return
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {

		ctx := stream.Context()

		reqID := HandleRequestID(ctx)

		ophid, _ := extractHeaderFromMetaData(ctx, "ophid")
		// add request id to logger
		addRequestIDToLogger(ctx, reqID)
		addOphIDToLogger(ctx, ophid)

		ctx = NewContext(ctx, reqID)
		ctx = NewContextWithOphID(ctx, ophid)

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = ctx

		err = handler(srv, wrapped)

		return
	}
}
