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

func TestHandler_OrderPay(t *testing.T) {
	testRequest := &pb.OrderPayRequest{
		// nolint:gosec
		OrderID: rand.Int63(),
	}

	expectResp := &pb.OrderPayResponse{}

	tests := []struct {
		name               string
		testRequest        *pb.OrderPayRequest
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedResp       *pb.OrderPayResponse
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.OrderPayHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", testRequest.OrderID),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderPayMock.
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
					model.OrderPayHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", testRequest.OrderID),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderPayMock.
					Expect(context.Background(), testRequest.OrderID).
					Return(model.ErrOrderIDNotFound)

				tc.mockTracer.StartMock.
					When(
						context.Background(),
						model.OrderPayHandler,
						trace.WithAttributes(
							attribute.Int64("OrderID", testRequest.OrderID),
							attribute.String("err", model.ErrOrderIDNotFound.Error()),
						),
					).Then(context.Background(), trace.SpanFromContext(context.Background()))
			},
			expectedStatusCode: codes.NotFound,
			expectedResp:       nil,
			expectedErr:        model.ErrOrderIDNotFound,
		},
		{
			name:        "err order already paid",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.OrderPayHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", testRequest.OrderID),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderPayMock.
					Expect(minimock.AnyContext, testRequest.OrderID).
					Return(model.ErrOrderAlreadyPay)

				tc.mockTracer.StartMock.
					When(
						context.Background(),
						model.OrderPayHandler,
						trace.WithAttributes(
							attribute.Int64("OrderID", testRequest.OrderID),
							attribute.String("err", model.ErrOrderAlreadyPay.Error()),
						),
					).Then(context.Background(), trace.SpanFromContext(context.Background()))
			},
			expectedStatusCode: codes.OK,
			expectedResp:       expectResp,
			expectedErr:        status.Error(codes.OK, model.ErrOrderAlreadyPay.Error()),
		},
		{
			name:        "err order already paid",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.OrderPayHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", testRequest.OrderID),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderPayMock.
					Expect(minimock.AnyContext, testRequest.OrderID).
					Return(model.ErrOrderStatusNotAwaitingPayment)

				tc.mockTracer.StartMock.
					When(
						context.Background(),
						model.OrderPayHandler,
						trace.WithAttributes(
							attribute.Int64("OrderID", testRequest.OrderID),
							attribute.String("err", model.ErrOrderStatusNotAwaitingPayment.Error()),
						),
					).Then(context.Background(), trace.SpanFromContext(context.Background()))
			},
			expectedStatusCode: codes.FailedPrecondition,
			expectedResp:       nil,
			expectedErr:        status.Error(codes.FailedPrecondition, model.ErrOrderStatusNotAwaitingPayment.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			resp, err := tc.server.OrderPay(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
