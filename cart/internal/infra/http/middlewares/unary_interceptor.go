package middlewares

import (
	"context"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryInterceptor ...
func UnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	metrics.IncRequestCount(method, model.TypeExternal)
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	duration := time.Since(start)
	code := status.Code(err)

	metrics.RequestDuration(method, code, model.TypeExternal, duration)

	return err
}
