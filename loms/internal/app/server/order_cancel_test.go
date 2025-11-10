package server

import (
	"context"
	"math/rand"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/pkg/api/v1"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_OrderCancel(t *testing.T) {
	testRequest := &pb.OrderCancelRequest{
		// nolint:gosec
		OrderID: rand.Int63(),
	}

	expectResp := &pb.OrderCancelResponse{}

	tests := []struct {
		name               string
		testRequest        *pb.OrderCancelRequest
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedResp       *pb.OrderCancelResponse
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.OrderCancelHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", testRequest.OrderID),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderCancelMock.
					Expect(minimock.AnyContext, testRequest.OrderID).
					Return(nil)
			},
			expectedStatusCode: codes.OK,
			expectedResp:       expectResp,
			expectedErr:        nil,
		},
		{
			name:        "err not found order by id",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.OrderCancelHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", testRequest.OrderID),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderCancelMock.
					Expect(minimock.AnyContext, testRequest.OrderID).
					Return(model.ErrOrderCancelNotFound)

				tc.mockTracer.StartMock.
					When(
						context.Background(),
						model.OrderCancelHandler,
						trace.WithAttributes(
							attribute.Int64("OrderID", testRequest.OrderID),
							attribute.String("err", model.ErrOrderCancelNotFound.Error()),
						),
					).Then(context.Background(), trace.SpanFromContext(context.Background()))
			},
			expectedStatusCode: codes.NotFound,
			expectedResp:       nil,
			expectedErr:        status.Error(codes.NotFound, model.ErrOrderCancelNotFound.Error()),
		},
		{
			name:        "err order already canceled",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.OrderCancelHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", testRequest.OrderID),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderCancelMock.
					Expect(minimock.AnyContext, testRequest.OrderID).
					Return(model.ErrOrderAlreadyCanceled)

				tc.mockTracer.StartMock.
					When(
						context.Background(),
						model.OrderCancelHandler,
						trace.WithAttributes(
							attribute.Int64("OrderID", testRequest.OrderID),
							attribute.String("err", model.ErrOrderAlreadyCanceled.Error()),
						),
					).Then(context.Background(), trace.SpanFromContext(context.Background()))
			},
			expectedStatusCode: codes.OK,
			expectedResp:       expectResp,
			expectedErr:        status.Error(codes.OK, model.ErrOrderAlreadyCanceled.Error()),
		},
		{
			name:        "err order status failed or paid",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.OrderCancelHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", testRequest.OrderID),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderCancelMock.
					Expect(minimock.AnyContext, testRequest.OrderID).
					Return(model.ErrOrderStatusFailedOrPaid)

				tc.mockTracer.StartMock.
					When(
						context.Background(),
						model.OrderCancelHandler,
						trace.WithAttributes(
							attribute.Int64("OrderID", testRequest.OrderID),
							attribute.String("err", model.ErrOrderStatusFailedOrPaid.Error()),
						),
					).Then(context.Background(), trace.SpanFromContext(context.Background()))
			},
			expectedStatusCode: codes.FailedPrecondition,
			expectedResp:       nil,
			expectedErr:        status.Error(codes.FailedPrecondition, model.ErrOrderStatusFailedOrPaid.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			resp, err := tc.server.OrderCancel(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
