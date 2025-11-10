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

func TestService_GetStocksBySku(t *testing.T) {
	// nolint:gosec
	testSku := rand.Int63()
	// nolint:gosec
	expectFreeStock := rand.Uint32()

	tests := []struct {
		name               string
		testRequest        int64
		setupMock          func(tc testComponent)
		expectResponce     uint32
		expectedStatusCode codes.Code
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testSku,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:GetStocksBySku",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.UseMasterMock.
					Expect(model.RequestStock).
					Return(true)
				tc.mockRepo.GetFreeStocksBySkuMasterMock.
					Expect(minimock.AnyContext, testSku).
					Return(expectFreeStock, nil)

			},
			expectResponce:     expectFreeStock,
			expectedStatusCode: codes.OK,
			expectedErr:        nil,
		},
		{
			name:        "err not found stock by sku",
			testRequest: testSku,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					"LomsService:GetStocksBySku",
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.UseMasterMock.
					Expect(model.RequestStock).
					Return(true)
				tc.mockRepo.GetFreeStocksBySkuMasterMock.
					Expect(context.Background(), testSku).
					Return(model.ErrorStockCount, model.ErrStockSkuNotFound)
			},
			expectResponce:     model.ErrorStockCount,
			expectedStatusCode: codes.NotFound,
			expectedErr:        model.ErrStockSkuNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			freeCount, err := tc.service.GetStocksBySku(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectResponce, freeCount)
			assert.Equal(t, tt.expectedErr, err)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
