package server

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestHandler_DeleteItem(t *testing.T) {
	const testURL = "/user/{user_id}/cart/{sku_id}"

	testData := model.RequestData{
		// nolint:gosec
		UserID: rand.Int63(),
		Sku:    int64(1076963),
	}

	tests := []struct {
		name           string
		testData       model.RequestData
		testBody       string
		setupMock      func(tc testComponent, mockData model.RequestData)
		expectedStatus int
		expectedBody   string
		expectedResp   error
	}{
		{
			name:     "success",
			testData: testData,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.DeleteItemURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
							attribute.Int64("Sku", testData.Sku),
						),
					).
					Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.DeleteItemMock.
					Expect(minimock.AnyContext, mockData).
					Return(model.ErrNoContent)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   fmt.Sprintf("{\"Message\":\"card %d deleted successfully\"}\n", testData.Sku),
			expectedResp:   nil,
		},
		{
			name:     "err delete item",
			testData: testData,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.DeleteItemURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
							attribute.Int64("Sku", testData.Sku),
						),
					).
					Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.DeleteItemMock.
					Expect(minimock.AnyContext, mockData).
					Return(errors.New("test"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "{\"Message\":\"test\"}\n",
			expectedResp:   nil,
		},
		{
			name:     "delete item status OK",
			testData: testData,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.DeleteItemURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
							attribute.Int64("Sku", testData.Sku),
						),
					).
					Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

				tc.mock.DeleteItemMock.
					Expect(minimock.AnyContext, mockData).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
			expectedResp:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tc := setupTest(t)
			tt.setupMock(tc, testData)

			// Execute
			req := httptest.NewRequest(http.MethodDelete, testURL, nil)
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("sku_id", fmt.Sprintf("%d", testData.Sku))
			req.SetPathValue("user_id", fmt.Sprintf("%d", testData.UserID))

			w := httptest.NewRecorder()
			tc.server.DeleteItem(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			// Verify
			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}

}
