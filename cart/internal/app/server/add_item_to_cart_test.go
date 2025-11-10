package server

import (
	"bytes"
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

func TestHandler_AddItem(t *testing.T) {
	const (
		testURL = "/user/{user_id}/cart/{sku_id}"
	)

	testData := model.RequestData{
		// nolint:gosec
		UserID: rand.Int63(),
		Sku:    int64(1076963),
		// nolint:gosec
		Count: rand.Uint32(),
	}

	testBody := fmt.Sprintf(`{"count":%d}`, testData.Count)

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
			testBody: testBody,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.AddItemURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
							attribute.Int64("Sku", testData.Sku),
							attribute.Int64("Count", int64(testData.Count)),
						),
					).
					Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

				tc.mock.AddItemMock.
					Expect(minimock.AnyContext, mockData).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "card add successfully",
			expectedResp:   nil,
		},
		{
			name:     "err 'proudct not found' add item",
			testData: testData,
			testBody: testBody,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.AddItemURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
							attribute.Int64("Sku", testData.Sku),
							attribute.Int64("Count", int64(testData.Count)),
						),
					).
					Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.AddItemMock.
					Expect(minimock.AnyContext, mockData).
					Return(model.ErrProductNotFound)
			},
			expectedStatus: http.StatusPreconditionFailed,
			expectedBody:   fmt.Sprintf("{\"Message\":\"%s\"}\n", model.ErrSkuNotExists),
			expectedResp:   model.ErrProductNotFound,
		},
		{
			name:     "err 'many request' add item",
			testData: testData,
			testBody: testBody,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.AddItemURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
							attribute.Int64("Sku", testData.Sku),
							attribute.Int64("Count", int64(testData.Count)),
						),
					).
					Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.AddItemMock.
					Expect(minimock.AnyContext, mockData).
					Return(model.ErrManyRequest)
			},
			expectedStatus: http.StatusTooManyRequests,
			expectedBody:   fmt.Sprintf("{\"Message\":\"%s\"}\n", model.ErrSkuNotExists),
			expectedResp:   model.ErrManyRequest,
		},
		{
			name:     "err internal add item",
			testData: testData,
			testBody: testBody,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.AddItemURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
							attribute.Int64("Sku", testData.Sku),
							attribute.Int64("Count", int64(testData.Count)),
						),
					).
					Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.AddItemMock.
					Expect(minimock.AnyContext, mockData).
					Return(errors.New("test"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   fmt.Sprintf("{\"Message\":\"%s\"}\n", "test"),
			expectedResp:   errors.New("test"),
		},
		{
			name:     "err added more item than in stock",
			testData: testData,
			testBody: testBody,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.AddItemURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
							attribute.Int64("Sku", testData.Sku),
							attribute.Int64("Count", int64(testData.Count)),
						),
					).
					Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

				tc.mock.AddItemMock.
					Expect(minimock.AnyContext, mockData).
					Return(model.ErrAddedMoreItemThanInStock)
			},
			expectedStatus: http.StatusPreconditionFailed,
			expectedBody:   fmt.Sprintf("{\"Message\":\"%s\"}\n", model.ErrAddedMoreItemThanInStock.Error()),
			expectedResp:   model.ErrAddedMoreItemThanInStock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tc := setupTest(t)
			if tc.mock == nil {
				t.Fatal("serviceMock is nil")
			}
			tt.setupMock(tc, testData)

			// Execute
			reader := bytes.NewReader([]byte(testBody))
			req := httptest.NewRequest(http.MethodPost, testURL, reader)
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("sku_id", fmt.Sprintf("%d", testData.Sku))
			req.SetPathValue("user_id", fmt.Sprintf("%d", testData.UserID))

			w := httptest.NewRecorder()
			tc.server.AddItem(w, req)

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

func TestParseRequest_err(t *testing.T) {

	const testURL = "/user/{user_id}/cart/{sku_id}"

	testData := model.RequestData{
		// nolint:gosec
		UserID: rand.Int63(),
		Sku:    int64(1076963),
	}

	t.Run("err parse request", func(t *testing.T) {
		reader := bytes.NewReader([]byte(""))
		req := httptest.NewRequest(http.MethodPost, testURL, reader)
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("sku_id", fmt.Sprintf("%d", testData.Sku))
		req.SetPathValue("user_id", fmt.Sprintf("%d", testData.UserID))
		_, err := parseRequest(req, int(model.ValidateFull))
		require.Error(t, err, errors.New(model.ErrCounItemsMoreThanZero))
	})
}
