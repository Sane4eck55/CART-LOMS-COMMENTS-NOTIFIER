package service

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestService_OrderCreate(t *testing.T) {
	// nolint:gosec
	expectOrderID := rand.Int63()

	testRequest := model.Order{
		// nolint:gosec
		UserID: rand.Int63(),
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

	tests := []struct {
		name               string
		testRequest        model.Order
		setupMock          func(tc testComponent)
		expectResponce     int64
		expectedStatusCode codes.Code
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderCreate",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				testRequest.Items = sortItems(testRequest.Items)

				tc.mockRepo.CreateOrderMock.
					Expect(minimock.AnyContext, testRequest).
					Return(expectOrderID, nil)
				tc.mockRepo.ReserveMock.
					When(minimock.AnyContext, testRequest.Items).
					Then(nil)
				tc.mockRepo.SetStatusOrderMock.
					Expect(minimock.AnyContext, expectOrderID, model.StatusOrderAwaitingPayment).
					Return(nil)
			},
			expectResponce:     expectOrderID,
			expectedStatusCode: codes.OK,
			expectedErr:        nil,
		},
		{
			name:        "err no stock for reserve",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderCreate",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				testRequest.Items = sortItems(testRequest.Items)

				tc.mockRepo.CreateOrderMock.
					Expect(minimock.AnyContext, testRequest).
					Return(expectOrderID, nil)
				tc.mockRepo.ReserveMock.
					When(minimock.AnyContext, testRequest.Items).
					Then(model.ErrNoStockForReserve)
				tc.mockRepo.SetStatusOrderMock.
					Expect(minimock.AnyContext, expectOrderID, model.StatusOrderFailed).
					Return(nil)
			},
			expectResponce:     expectOrderID,
			expectedStatusCode: codes.FailedPrecondition,
			expectedErr:        model.ErrNoStockForReserve,
		},
		{
			name:        "err stock info not found",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderCreate",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				testRequest.Items = sortItems(testRequest.Items)

				tc.mockRepo.CreateOrderMock.
					Expect(minimock.AnyContext, testRequest).
					Return(expectOrderID, nil)
				tc.mockRepo.ReserveMock.
					When(minimock.AnyContext, testRequest.Items).
					Then(model.ErrStockInfoNotFound)
				tc.mockRepo.SetStatusOrderMock.
					Expect(minimock.AnyContext, expectOrderID, model.StatusOrderFailed).
					Return(nil)
			},
			expectResponce:     expectOrderID,
			expectedStatusCode: codes.FailedPrecondition,
			expectedErr:        model.ErrStockInfoNotFound,
		},
		{
			name:        "err create order",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderCreate",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				testRequest.Items = sortItems(testRequest.Items)

				tc.mockRepo.CreateOrderMock.
					Expect(minimock.AnyContext, testRequest).
					Return(model.ErrorOrderID, errors.New("test"))

			},
			expectResponce:     model.ErrorOrderID,
			expectedStatusCode: codes.Unknown,
			expectedErr:        errors.New("test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			orderID, err := tc.service.OrderCreate(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectResponce, orderID)
			assert.Equal(t, tt.expectedErr, err)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
