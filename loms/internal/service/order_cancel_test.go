package service

import (
	"context"
	"math/rand"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestService_OrderCancel(t *testing.T) {
	// nolint:gosec
	testOrderID := rand.Int63()

	expectResponce := model.OrderInfo{
		// nolint:gosec
		UserID: rand.Int63(),
		Status: model.StatusOrderAwaitingPayment,
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
		testRequest        int64
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testOrderID,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderCancel",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.GetInfoByOrderIDMasterMock.
					Expect(minimock.AnyContext, testOrderID).
					Return(&expectResponce, nil)
				tc.mockRepo.ReserveCancelMock.
					When(minimock.AnyContext, expectResponce.Items[0]).
					Then(nil)
				tc.mockRepo.ReserveCancelMock.
					When(minimock.AnyContext, expectResponce.Items[1]).
					Then(nil)
				tc.mockRepo.SetStatusOrderMock.
					Expect(minimock.AnyContext, testOrderID, model.StatusOrderCancelled).
					Return(nil)
			},
			expectedStatusCode: codes.OK,
			expectedErr:        nil,
		},
		{
			name:        "err not found order by id",
			testRequest: testOrderID,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderCancel",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.GetInfoByOrderIDMasterMock.
					Expect(minimock.AnyContext, testOrderID).
					Return(nil, model.ErrOrderPayNotFound)
			},
			expectedStatusCode: codes.NotFound,
			expectedErr:        model.ErrOrderCancelNotFound,
		},
		{
			name:        "err order already canceled",
			testRequest: testOrderID,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderCancel",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.GetInfoByOrderIDMasterMock.
					Expect(minimock.AnyContext, testOrderID).
					Return(&model.OrderInfo{
						Status: model.StatusOrderCancelled,
					}, nil)
			},
			expectedStatusCode: codes.OK,
			expectedErr:        model.ErrOrderAlreadyCanceled,
		},
		{
			name:        "err order status failed or paid",
			testRequest: testOrderID,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderCancel",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.GetInfoByOrderIDMasterMock.
					Expect(minimock.AnyContext, testOrderID).
					Return(&model.OrderInfo{
						Status: model.StatusOrderFailed,
					}, nil)

			},
			expectedStatusCode: codes.FailedPrecondition,
			expectedErr:        model.ErrOrderStatusFailedOrPaid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			err := tc.service.OrderCancel(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
