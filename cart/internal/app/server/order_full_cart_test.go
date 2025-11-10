package server

import (
	"context"
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

func TestHandler_OrderFullCart(t *testing.T) {
	const (
		testURL = "/checkout/{user_id}"
	)

	testData := model.RequestData{
		// nolint:gosec
		UserID: rand.Int63(),
	}

	expectItems := []model.Item{
		{
			// nolint:gosec
			Sku:   rand.Int63(),
			Name:  "item",
			Count: 1,
			Price: 100,
		},
		{
			// nolint:gosec
			Sku:   rand.Int63(),
			Name:  "item2",
			Count: 2,
			Price: 30,
		},
		{
			// nolint:gosec
			Sku:   rand.Int63(),
			Name:  "item3",
			Count: 3,
			Price: 25,
		},
	}

	expectGetItemsFromCartResponce := model.GetItemsFromCartResponce{
		Items:      expectItems,
		TotalPrice: countTotalPrice(expectItems),
	}
	// nolint:gosec
	expectOrderID := rand.Int63()

	tests := []struct {
		name           string
		testData       model.RequestData
		setupMock      func(tc testComponent, mockData model.RequestData)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "success",
			testData: testData,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					When(
						context.Background(),
						model.OrderFullCartURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
						),
					).
					Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))

				tc.mock.GetItemsFromCartMock.
					Expect(minimock.AnyContext, mockData).
					Return(&expectGetItemsFromCartResponce, nil)

				tc.mock.OrderCreateMock.
					Expect(minimock.AnyContext, testData.UserID, &expectGetItemsFromCartResponce).
					Return(expectOrderID, nil)

				tc.mock.DeleteItemsByUserIDMock.
					Expect(minimock.AnyContext, testData).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   fmt.Sprintf("{\"order_id\":%d}\n", expectOrderID),
		},
		{
			name:     "err cart empty",
			testData: testData,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.OrderFullCartURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
						),
					).
					Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mock.GetItemsFromCartMock.
					Expect(minimock.AnyContext, mockData).
					Return(nil, model.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   fmt.Sprintf("{\"Message\":\"%s\"}\n", model.ErrCartEmpty.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tc := setupTest(t)
			tt.setupMock(tc, testData)

			// Execute
			req := httptest.NewRequest(http.MethodPost, testURL, nil)
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("user_id", fmt.Sprintf("%d", testData.UserID))

			w := httptest.NewRecorder()
			tc.server.OrderFullCart(w, req)

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

func countTotalPrice(items []model.Item) uint32 {
	var totalPrice uint32
	for _, item := range items {
		totalPrice += (item.Price * item.Count)
	}

	return totalPrice
}
