package middlewares

import (
	"context"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/infra/metrics"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryInterceptor ...
func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	method := info.FullMethod

	metrics.IncRequestCount(method, model.TypeExternal)
	start := time.Now()

	resp, err = handler(ctx, req)

	duration := time.Since(start)
	code := status.Code(err)

	metrics.RequestDuration(method, code.String(), model.TypeExternal, duration)

	return resp, err
}
