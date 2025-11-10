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

func TestService_OrderInfo(t *testing.T) {
	// nolint:gosec
	testOrderID := rand.Int63()

	expectOrderInfo := model.OrderInfo{
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

	tests := []struct {
		name               string
		testRequest        int64
		setupMock          func(tc testComponent)
		expectResponce     *model.OrderInfo
		expectedStatusCode codes.Code
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testOrderID,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderInfo",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.UseMasterMock.
					Expect(model.RequestOrder).
					Return(true)
				tc.mockRepo.GetInfoByOrderIDMasterMock.
					Expect(minimock.AnyContext, testOrderID).
					Return(&expectOrderInfo, nil)
			},
			expectResponce:     &expectOrderInfo,
			expectedStatusCode: codes.OK,
			expectedErr:        nil,
		},
		{
			name:        "err not found info by order id",
			testRequest: testOrderID,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:OrderInfo",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.UseMasterMock.
					Expect(model.RequestOrder).
					Return(true)
				tc.mockRepo.GetInfoByOrderIDMasterMock.
					Expect(minimock.AnyContext, testOrderID).
					Return(nil, model.ErrOrderIDNotFound)
			},
			expectResponce:     nil,
			expectedStatusCode: codes.NotFound,
			expectedErr:        model.ErrOrderIDNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			orderID, err := tc.service.OrderInfo(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectResponce, orderID)
			assert.Equal(t, tt.expectedErr, err)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
