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

func TestHandler_OrderInfo(t *testing.T) {
	testRequest := &pb.OrderInfoRequest{
		// nolint:gosec
		OrderID: rand.Int63(),
	}

	testOrderInfo := model.OrderInfo{
		// nolint:gosec
		UserID: rand.Int63(),
		Status: model.StatusOrderPaid,
		Items: []model.Item{
			{
				// nolint:gosec
				Sku: rand.Int63(),
				// nolint:gosec
				Count: rand.Uint32(),
			},
			{
				// nolint:gosec
				Sku: rand.Int63(),
				// nolint:gosec
				Count: rand.Uint32(),
			},
		},
	}
	expectResp := orderInfoToOrderInfoResponse(testOrderInfo)

	tests := []struct {
		name               string
		testRequest        *pb.OrderInfoRequest
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedResp       *pb.OrderInfoResponse
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.OrderInfoHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", testRequest.OrderID),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderInfoMock.
					Expect(minimock.AnyContext, testRequest.OrderID).
					Return(&testOrderInfo, nil)
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
					model.OrderInfoHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", testRequest.OrderID),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderInfoMock.
					Expect(minimock.AnyContext, testRequest.OrderID).
					Return(nil, model.ErrOrderIDNotFound)

				tc.mockTracer.StartMock.
					When(
						context.Background(),
						model.OrderInfoHandler,
						trace.WithAttributes(
							attribute.Int64("OrderID", testRequest.OrderID),
							attribute.String("err", model.ErrOrderIDNotFound.Error()),
						),
					).Then(context.Background(), trace.SpanFromContext(context.Background()))
			},
			expectedStatusCode: codes.NotFound,
			expectedResp:       nil,
			expectedErr:        status.Error(codes.NotFound, model.ErrOrderIDNotFound.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			resp, err := tc.server.OrderInfo(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
