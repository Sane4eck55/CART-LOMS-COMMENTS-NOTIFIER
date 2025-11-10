package server

import (
	"context"
	"math/rand"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/pkg/api/v1"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
)

func TestHandler_OrderCreate(t *testing.T) {
	testItems := []*pb.Item{
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
	}

	testRequest := &pb.OrderCreateRequest{
		// nolint:gosec
		UserID: rand.Int63(),
		Items:  testItems,
	}
	// nolint:gosec
	expectOrderID := rand.Int63()
	expectResp := &pb.OrderCreateResponse{
		OrderID: expectOrderID,
	}

	tests := []struct {
		name               string
		testRequest        *pb.OrderCreateRequest
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedResp       *pb.OrderCreateResponse
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				var attrs []attribute.KeyValue
				for _, item := range testItems {
					attrs = append(attrs,
						attribute.Int64("UserID", testRequest.GetUserID()),
						attribute.Int64("Sku", item.GetSku()),
						attribute.Int64("Count", int64(item.GetCount())),
					)
				}

				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.OrderCreateHandler,
					trace.WithAttributes(
						attrs...,
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderCreateMock.
					Expect(context.Background(), preparedInputOrderCreate(testRequest)).
					Return(expectOrderID, nil)
			},
			expectedStatusCode: codes.OK,
			expectedResp:       expectResp,
			expectedErr:        nil,
		},
		{
			name:        "err no stock for reserve",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				var attrs []attribute.KeyValue
				for _, item := range testItems {
					attrs = append(attrs,
						attribute.Int64("UserID", testRequest.GetUserID()),
						attribute.Int64("Sku", item.GetSku()),
						attribute.Int64("Count", int64(item.GetCount())),
					)
				}

				tc.mockTracer.StartMock.
					Expect(
						context.Background(),
						model.OrderCreateHandler,
						trace.WithAttributes(
							attrs...,
						),
					).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.OrderCreateMock.
					Expect(minimock.AnyContext, preparedInputOrderCreate(testRequest)).
					Return(expectOrderID, model.ErrNoStockForReserve)

				tc.mockTracer.StartMock.
					When(
						context.Background(),
						model.OrderCreateHandler,
						trace.WithAttributes(
							attribute.Int64("UserID", testRequest.GetUserID()),
							attribute.String("err", model.ErrNoStockForReserve.Error()),
						),
					).Then(context.Background(), trace.SpanFromContext(context.Background()))
			},
			expectedStatusCode: codes.FailedPrecondition,
			expectedResp:       expectResp,
			expectedErr:        status.Error(codes.FailedPrecondition, model.ErrNoStockForReserve.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			resp, err := tc.server.OrderCreate(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}

func TestHandler_preparedInputOrderCreate(t *testing.T) {
	// nolint:gosec
	testUserID := rand.Int63()
	// nolint:gosec
	testSku1 := rand.Int63()
	// nolint:gosec
	testSku2 := rand.Int63()
	// nolint:gosec
	testCount1 := rand.Uint32()
	// nolint:gosec
	testCount2 := rand.Uint32()

	tests := []struct {
		name         string
		testRequest  *pb.OrderCreateRequest
		expectedResp model.Order
	}{
		{
			name: "success",
			testRequest: &pb.OrderCreateRequest{
				UserID: testUserID,
				Items: []*pb.Item{
					{
						Sku:   testSku1,
						Count: testCount1,
					},
					{
						Sku:   testSku2,
						Count: testCount2,
					},
				},
			},
			expectedResp: model.Order{
				UserID: testUserID,
				Items: []model.Item{
					{
						Sku:   testSku1,
						Count: testCount1,
					},
					{
						Sku:   testSku2,
						Count: testCount2,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := preparedInputOrderCreate(tt.testRequest)
			assert.Equal(t, res, tt.expectedResp)
		})
	}
}
