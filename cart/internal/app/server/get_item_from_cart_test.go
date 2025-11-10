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

func TestGetItemsByUserID(t *testing.T) {
	t.Parallel()
	const testURL = "/user/{user_id}/cart"

	t.Run("successful get items", func(t *testing.T) {
		data := model.RequestData{
			// nolint:gosec
			UserID: rand.Int63(),
		}

		item := model.Item{
			Sku:   1,
			Name:  "item",
			Count: 1,
			Price: 1,
		}
		item2 := model.Item{
			Sku:   2,
			Name:  "item2",
			Count: 2,
			Price: 2,
		}

		totalPrice := item.Price + item2.Price

		items := []model.Item{item, item2}
		tc := setupTest(t)
		tc.tracer.StartMock.
			Expect(
				context.Background(),
				model.GetItemsByUserIDURL,
				trace.WithAttributes(
					attribute.Int64("UserID", data.UserID),
				),
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		tc.mock.GetItemsFromCartMock.
			Expect(minimock.AnyContext, data).
			Return(
				&model.GetItemsFromCartResponce{
					Items:      items,
					TotalPrice: totalPrice,
				}, nil)

		req := httptest.NewRequest(http.MethodGet, testURL, nil)
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("user_id", fmt.Sprintf("%d", data.UserID))

		w := httptest.NewRecorder()
		tc.server.GetItemsByUserID(w, req)

		res := w.Result()
		defer func() {
			err := res.Body.Close()
			require.NoError(t, err)
		}()

		require.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("err 'not found' get items", func(t *testing.T) {
		t.Parallel()
		data := model.RequestData{
			UserID: 1,
		}

		tc := setupTest(t)
		tc.tracer.StartMock.
			Expect(
				context.Background(),
				model.GetItemsByUserIDURL,
				trace.WithAttributes(
					attribute.Int64("UserID", data.UserID),
				),
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		tc.mock.GetItemsFromCartMock.
			Expect(minimock.AnyContext, data).
			Return(nil, model.ErrNotFound)

		req := httptest.NewRequest(http.MethodGet, testURL, nil)
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("user_id", fmt.Sprintf("%d", data.UserID))

		w := httptest.NewRecorder()
		tc.server.GetItemsByUserID(w, req)

		res := w.Result()
		defer func() {
			err := res.Body.Close()
			require.NoError(t, err)
		}()

		require.Equal(t, http.StatusNotFound, res.StatusCode)

	})

	t.Run("err  get items", func(t *testing.T) {
		t.Parallel()
		data := model.RequestData{
			UserID: 1,
		}

		tc := setupTest(t)
		tc.tracer.StartMock.
			Expect(
				context.Background(),
				model.GetItemsByUserIDURL,
				trace.WithAttributes(
					attribute.Int64("UserID", data.UserID),
				),
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		tc.mock.GetItemsFromCartMock.
			Expect(minimock.AnyContext, data).
			Return(nil, errors.New("test"))

		req := httptest.NewRequest(http.MethodGet, testURL, nil)
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("user_id", fmt.Sprintf("%d", data.UserID))

		w := httptest.NewRecorder()
		tc.server.GetItemsByUserID(w, req)

		res := w.Result()
		defer func() {
			err := res.Body.Close()
			require.NoError(t, err)
		}()

		require.Equal(t, http.StatusNoContent, res.StatusCode)

	})
}

func TestHandler_GetItemsByUserID(t *testing.T) {
	const testURL = "/user/{user_id}/cart"
	testData := model.RequestData{
		// nolint:gosec
		UserID: rand.Int63(),
	}

	testItem := model.Item{
		Sku:   1,
		Name:  "item",
		Count: 1,
		Price: 1,
	}
	testItem2 := model.Item{
		Sku:   2,
		Name:  "item2",
		Count: 2,
		Price: 2,
	}

	totalPrice := testItem.Price + (testItem2.Price * testItem2.Count)
	testItems := []model.Item{testItem, testItem2}

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
						model.GetItemsByUserIDURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
						),
					).
					Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

				tc.mock.GetItemsFromCartMock.
					Expect(minimock.AnyContext, mockData).
					Return(&model.GetItemsFromCartResponce{
						Items:      testItems,
						TotalPrice: totalPrice,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			//TO DO: сделать унифицированый метод, если изменится слайс items по кол-ву, то expectedBody будет другой
			expectedBody: fmt.Sprintf("{\"items\":[{\"sku\":%d,\"name\":\"%s\",\"count\":%d,\"price\":%d},{\"sku\":%d,\"name\":\"%s\",\"count\":%d,\"price\":%d}],\"total_price\":%d}\n",
				testItem.Sku, testItem.Name, testItem.Count, testItem.Price, testItem2.Sku, testItem2.Name, testItem2.Count, testItem2.Price, totalPrice),
			expectedResp: nil,
		},
		{
			name:     "err 'not found' get items",
			testData: testData,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.GetItemsByUserIDURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
						),
					).
					Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

				tc.mock.GetItemsFromCartMock.
					Expect(minimock.AnyContext, mockData).
					Return(nil, model.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "{\"Message\":\"not found\"}\n",
			expectedResp:   nil,
		},
		{
			name:     "err  get items",
			testData: testData,
			setupMock: func(tc testComponent, mockData model.RequestData) {
				tc.tracer.StartMock.
					Expect(
						context.Background(),
						model.GetItemsByUserIDURL,
						trace.WithAttributes(
							attribute.Int64("UserID", testData.UserID),
						),
					).
					Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

				tc.mock.GetItemsFromCartMock.
					Expect(minimock.AnyContext, mockData).
					Return(nil, errors.New("test"))
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "{\"Message\":\"test\"}\n",
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
			req.SetPathValue("user_id", fmt.Sprintf("%d", testData.UserID))

			w := httptest.NewRecorder()
			tc.server.GetItemsByUserID(w, req)

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
