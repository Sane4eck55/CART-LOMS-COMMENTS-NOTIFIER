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

func TestHandler_StocksInfo(t *testing.T) {
	testRequest := &pb.StocksInfoRequest{
		// nolint:gosec
		Sku: rand.Int63(),
	}

	expectResp := &pb.StocksInfoResponse{
		// nolint:gosec
		Count: rand.Uint32(),
	}

	tests := []struct {
		name               string
		testRequest        *pb.StocksInfoRequest
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedResp       *pb.StocksInfoResponse
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.StocksInfoHandler,
					trace.WithAttributes(
						attribute.Int64("Sku", testRequest.GetSku()),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.GetStocksBySkuMock.
					Expect(minimock.AnyContext, testRequest.Sku).
					Return(expectResp.Count, nil)
			},
			expectedStatusCode: codes.OK,
			expectedResp:       expectResp,
			expectedErr:        nil,
		},
		{
			name:        "err not found stock info",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mockTracer.StartMock.Expect(
					context.Background(),
					model.StocksInfoHandler,
					trace.WithAttributes(
						attribute.Int64("Sku", testRequest.GetSku()),
					),
				).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.GetStocksBySkuMock.
					Expect(minimock.AnyContext, testRequest.Sku).
					Return(model.ErrorStockCount, model.ErrStockSkuNotFound)

				tc.mockTracer.StartMock.
					When(
						context.Background(),
						model.StocksInfoHandler,
						trace.WithAttributes(
							attribute.Int64("Sku", testRequest.GetSku()),
							attribute.String("err", model.ErrStockSkuNotFound.Error()),
						),
					).Then(context.Background(), trace.SpanFromContext(context.Background()))
			},
			expectedStatusCode: codes.NotFound,
			expectedResp: &pb.StocksInfoResponse{
				Count: model.ErrorStockCount,
			},
			expectedErr: status.Error(codes.NotFound, model.ErrStockSkuNotFound.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			resp, err := tc.server.StocksInfo(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
