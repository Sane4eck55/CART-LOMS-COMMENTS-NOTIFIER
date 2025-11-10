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

func TestService_OrderPay(t *testing.T) {
	// nolint:gosec
	testOrderID := rand.Int63()

	expectOrderInfo := model.OrderInfo{
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
					"LomsService:OrderPay",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.GetInfoByOrderIDMasterMock.
					Expect(minimock.AnyContext, testOrderID).
					Return(&expectOrderInfo, nil)
				tc.mockRepo.ReserveRemoveMock.
					When(minimock.AnyContext, expectOrderInfo.Items[0]).
					Then(nil)
				tc.mockRepo.ReserveRemoveMock.
					When(minimock.AnyContext, expectOrderInfo.Items[1]).
					Then(nil)
				tc.mockRepo.SetStatusOrderMock.
					Expect(minimock.AnyContext, testOrderID, model.StatusOrderPaid).
					Return(nil)
			},
			expectedStatusCode: codes.OK,
			expectedErr:        nil,
		},
		{
			name:        "err not found info by order id",
			testRequest: testOrderID,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderPay",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.GetInfoByOrderIDMasterMock.
					Expect(context.Background(), testOrderID).
					Return(nil, nil)
			},
			expectedStatusCode: codes.NotFound,
			expectedErr:        model.ErrOrderPayNotFound,
		},
		{
			name:        "err status order is paid",
			testRequest: testOrderID,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderPay",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.GetInfoByOrderIDMasterMock.
					Expect(minimock.AnyContext, testOrderID).
					Return(&model.OrderInfo{
						UserID: expectOrderInfo.UserID,
						Status: model.StatusOrderPaid,
						Items:  expectOrderInfo.Items,
					}, nil)
			},
			expectedStatusCode: codes.OK,
			expectedErr:        model.ErrOrderAlreadyPay,
		},
		{
			name:        "err status order not AwaitingPayment",
			testRequest: testOrderID,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderPay",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.GetInfoByOrderIDMasterMock.
					Expect(context.Background(), testOrderID).
					Return(&model.OrderInfo{
						UserID: expectOrderInfo.UserID,
						Status: model.StatusOrderFailed,
						Items:  expectOrderInfo.Items,
					}, nil)
			},
			expectedStatusCode: codes.FailedPrecondition,
			expectedErr:        model.ErrOrderStatusNotAwaitingPayment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			err := tc.service.OrderPay(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
