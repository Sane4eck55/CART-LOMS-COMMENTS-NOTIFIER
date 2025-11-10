package service

import (
	"context"
	"math/rand"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/errgroup"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/goleak"
)

func TestService_GetItemsFromCart(t *testing.T) {
	testData := model.RequestData{
		// nolint:gosec
		UserID: rand.Int63(),
	}

	testRepoResp := []model.Cart{
		{
			SkuID: 1,
			// nolint:gosec
			Count: rand.Uint32(),
		},
		{
			SkuID: 2,
			// nolint:gosec
			Count: rand.Uint32(),
		},
	}

	testResponse := &model.GetItemsFromCartResponce{}

	testResponse.Items = append(testResponse.Items, model.Item{})

	safePriceSku, _ := SafeInt64ToUint32(100)

	tests := []struct {
		name         string
		testData     model.RequestData
		setupMock    func(tc testServiceComponent)
		expectedErr  error
		expectedResp *model.GetItemsFromCartResponce
	}{
		{
			name:     "success for one item",
			testData: testData,
			setupMock: func(tc testServiceComponent) {
				tc.mockTrace.StartMock.
					Expect(
						context.Background(),
						"CartService:GetItemsFromCart",
					).Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.GetItemsByUserIDMock.
					Expect(minimock.AnyContext, testData).
					Return([]model.Cart{testRepoResp[0]}, nil)

				ch := make(chan model.Cart, len(testRepoResp))
				g, gCtx := errgroup.WithContext(context.Background())
				for i := 0; i < countGoroutines; i++ {
					g.Go(func() error {
						for item := range ch {
							if err := tc.service.limiterPS.Wait(gCtx); err != nil {
								t.Logf("err : %v", err)
							}
							tc.mockPS.GetProductBySkuMock.
								Expect(minimock.AnyContext, item.SkuID).
								Return(&model.GetProductResponse{
									Name:  "test",
									Price: 100,
									Sku:   item.SkuID,
								}, nil)
						}
						return nil
					})
				}

				ch <- testRepoResp[0]

				close(ch)

				if err := g.Wait(); err != nil {
					t.Logf("g.Wait() : %v", err)
				}

			},
			expectedResp: &model.GetItemsFromCartResponce{
				Items: []model.Item{
					{
						Sku:   testRepoResp[0].SkuID,
						Name:  "test",
						Count: testRepoResp[0].Count,
						Price: safePriceSku,
					},
				},
				TotalPrice: safePriceSku * testRepoResp[0].Count,
			},
		},
		{
			name:     "success for two item",
			testData: testData,
			setupMock: func(tc testServiceComponent) {
				tc.mockTrace.StartMock.
					Expect(
						context.Background(),
						"CartService:GetItemsFromCart",
					).
					Return(context.Background(), trace.SpanFromContext(context.Background()))

				tc.mockRepo.GetItemsByUserIDMock.
					Expect(minimock.AnyContext, testData).
					Return(testRepoResp, nil)

				for _, item := range testRepoResp {

					tc.mockPS.GetProductBySkuMock.When(minimock.AnyContext, item.SkuID).
						Then(&model.GetProductResponse{
							Name:  "test",
							Price: 100,
							Sku:   item.SkuID,
						}, nil)
				}

				ch := make(chan model.Cart, len(testRepoResp))
				g, gCtx := errgroup.WithContext(context.Background())
				for i := 0; i < countGoroutines; i++ {
					g.Go(func() error {
						for item := range ch {
							if err := tc.service.limiterPS.Wait(gCtx); err != nil {
								t.Logf("err : %v", err)
								return err
							}
							_, err := tc.mockPS.GetProductBySku(minimock.AnyContext, item.SkuID)
							if err != nil {
								t.Logf("err : %v", err)
								return err
							}
						}
						return nil
					})
				}

				for _, item := range testRepoResp {
					ch <- item
				}

				close(ch)

				if err := g.Wait(); err != nil {
					t.Logf("g.Wait() : %v", err)
				}

			},
			expectedResp: &model.GetItemsFromCartResponce{
				Items: []model.Item{
					{
						Sku:   testRepoResp[0].SkuID,
						Name:  "test",
						Count: testRepoResp[0].Count,
						Price: safePriceSku,
					},
					{
						Sku:   testRepoResp[1].SkuID,
						Name:  "test",
						Count: testRepoResp[1].Count,
						Price: safePriceSku,
					},
				},
				TotalPrice: safePriceSku*testRepoResp[0].Count +
					safePriceSku*testRepoResp[1].Count,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			// Setup
			tc := setupTest(t)
			tt.setupMock(tc)

			result, err := tc.service.GetItemsFromCart(context.Background(), testData)
			require.NoError(t, err)
			assert.Len(t, result.Items, len(tt.expectedResp.Items))
			assert.Equal(t, result.TotalPrice, tt.expectedResp.TotalPrice)

		})
	}
}
