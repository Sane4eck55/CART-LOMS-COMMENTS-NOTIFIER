package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/goleak"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/service/mocks"
)

func TestAdd(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	item := model.RequestData{
		UserID: 1,
		Sku:    1,
		Count:  1,
	}

	item2 := model.RequestData{
		UserID: 1,
		Sku:    2,
		Count:  1,
	}

	t.Run("add/get/delBySku one item", func(t *testing.T) {
		// init mock
		tracer := mocks.NewTracerMock(t)
		tracer.StartMock.
			Expect(
				context.Background(),
				"CartRepo:Add",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		// init repo
		repo := NewInMemoryRepository(tracer)
		defer repo.Close()
		// preprocess data
		err := repo.Add(ctx, item)
		require.NoError(t, err)

		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:GetItemsByUserID",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		items, err := repo.GetItemsByUserID(ctx, item)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, item.Sku, items[0].SkuID)
		assert.Equal(t, item.Count, items[0].Count)

		//test process

		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:DeleteItemsBySku",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		err = repo.DeleteItemsBySku(ctx, item)
		require.Error(t, err, model.ErrNoContent)
		//check result test
		_, err = repo.GetItemsByUserID(ctx, item)
		require.Error(t, err, model.ErrNotFound)
	})

	t.Run("add/get/delByUser one item", func(t *testing.T) {
		// init mock
		tracer := mocks.NewTracerMock(t)
		tracer.StartMock.
			Expect(
				context.Background(),
				"CartRepo:Add",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		// init repo
		repo := NewInMemoryRepository(tracer)
		defer repo.Close()
		// preprocess data
		err := repo.Add(ctx, item)
		require.NoError(t, err)

		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:GetItemsByUserID",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		items, err := repo.GetItemsByUserID(ctx, item)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, item.Sku, items[0].SkuID)
		assert.Equal(t, item.Count, items[0].Count)
		//test process
		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:DeleteItemsBySku",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		err = repo.DeleteItemsBySku(ctx, item)
		require.Error(t, err, model.ErrNoContent)
		//check result test
		_, err = repo.GetItemsByUserID(ctx, item)
		require.Error(t, err, model.ErrNotFound)

	})

	t.Run("add/get/delBySku one item twice", func(t *testing.T) {
		// init mock
		tracer := mocks.NewTracerMock(t)
		tracer.StartMock.
			Expect(
				context.Background(),
				"CartRepo:Add",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		// init repo
		repo := NewInMemoryRepository(tracer)
		defer repo.Close()
		// preprocess data
		err := repo.Add(ctx, item)
		require.NoError(t, err)
		err = repo.Add(ctx, item)
		require.NoError(t, err)

		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:GetItemsByUserID",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		items, err := repo.GetItemsByUserID(ctx, item)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, item.Sku, items[0].SkuID)
		assert.Equal(t, uint32(2), items[0].Count)
		//test process
		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:DeleteItemsBySku",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		err = repo.DeleteItemsBySku(ctx, item)
		require.Error(t, err, model.ErrNoContent)
		//check result test
		_, err = repo.GetItemsByUserID(ctx, item)
		require.Error(t, err, model.ErrNotFound)
	})

	t.Run("add/get/delByUser one item twice", func(t *testing.T) {
		// init mock
		tracer := mocks.NewTracerMock(t)
		tracer.StartMock.
			Expect(
				context.Background(),
				"CartRepo:Add",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:GetItemsByUserID",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:DeleteAllItemsFromCart",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		// init repo
		repo := NewInMemoryRepository(tracer)
		defer repo.Close()
		// preprocess data
		err := repo.Add(ctx, item)
		require.NoError(t, err)
		err = repo.Add(ctx, item)
		require.NoError(t, err)

		items, err := repo.GetItemsByUserID(ctx, item)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, item.Sku, items[0].SkuID)
		assert.Equal(t, uint32(2), items[0].Count)
		//test process
		err = repo.DeleteAllItemsFromCart(ctx, item)
		require.Error(t, err, model.ErrNoContent)
		//check result test
		_, err = repo.GetItemsByUserID(ctx, item)
		require.Error(t, err, model.ErrNotFound)

	})

	t.Run("add/get/delBySku two item", func(t *testing.T) {
		// init mock
		tracer := mocks.NewTracerMock(t)
		tracer.StartMock.
			Expect(
				context.Background(),
				"CartRepo:Add",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:GetItemsByUserID",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:DeleteItemsBySku",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		// init repo
		repo := NewInMemoryRepository(tracer)
		defer repo.Close()
		// preprocess data
		err := repo.Add(ctx, item)
		require.NoError(t, err)
		err = repo.Add(ctx, item2)
		require.NoError(t, err)

		items, err := repo.GetItemsByUserID(ctx, item)
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, item.Sku, items[0].SkuID)
		assert.Equal(t, item.Count, items[0].Count)
		assert.Equal(t, item2.Sku, items[1].SkuID)
		assert.Equal(t, item2.Count, items[1].Count)
		//test process
		err = repo.DeleteItemsBySku(ctx, item)
		require.Error(t, err, model.ErrNoContent)
		//check result test
		items, err = repo.GetItemsByUserID(ctx, item2)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, item2.Sku, items[0].SkuID)
		assert.Equal(t, item2.Count, items[0].Count)
	})

	t.Run("add/get/delByUser two item", func(t *testing.T) {
		// init mock
		tracer := mocks.NewTracerMock(t)
		tracer.StartMock.
			Expect(
				context.Background(),
				"CartRepo:Add",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:GetItemsByUserID",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		tracer.StartMock.
			When(
				minimock.AnyContext,
				"CartRepo:DeleteAllItemsFromCart",
			).
			Then(minimock.AnyContext, trace.SpanFromContext(context.Background()))
		// init repo
		repo := NewInMemoryRepository(tracer)
		defer repo.Close()
		// preprocess data
		err := repo.Add(ctx, item)
		require.NoError(t, err)
		err = repo.Add(ctx, item2)
		require.NoError(t, err)

		items, err := repo.GetItemsByUserID(ctx, item)
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, item.Sku, items[0].SkuID)
		assert.Equal(t, item.Count, items[0].Count)
		assert.Equal(t, item2.Sku, items[1].SkuID)
		assert.Equal(t, item2.Count, items[1].Count)
		//test process
		err = repo.DeleteAllItemsFromCart(ctx, item)
		require.Error(t, err, model.ErrNoContent)
		//check result test
		_, err = repo.GetItemsByUserID(ctx, item)
		require.Error(t, err, model.ErrNotFound)
	})

}

func TestRepo_Goroutine(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var (
		countGoroutine = 10
	)

	leaks := goleak.Find()
	logger.Infow(fmt.Sprintf("leaks : %v", leaks))

	t.Run("repo_add", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		// init mock
		tracer := mocks.NewTracerMock(t)

		repo := NewInMemoryRepository(tracer)
		defer repo.Close()

		var (
			userID int64 = 100
		)

		tracer.StartMock.
			Expect(
				context.Background(),
				"CartRepo:Add",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		for i := 0; i < countGoroutine; i++ {
			wg.Add(1)
			go func() {
				// nolint:gosec
				sku := rand.Int63()
				// nolint:gosec
				count := rand.Uint32()
				defer wg.Done()

				err := repo.Add(ctx, model.RequestData{
					UserID: userID,
					Sku:    sku,
					Count:  count,
				})
				if err != nil {
					t.Errorf("Add error: %v", err)
				}
			}()
		}

		wg.Wait()
		tracer.StartMock.
			Expect(
				minimock.AnyContext,
				"CartRepo:GetItemsByUserID",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		items, err := repo.GetItemsByUserID(ctx, model.RequestData{UserID: userID})
		if err != nil {
			t.Errorf("GetItemsByUserID error: %v", err)
		} else if len(items) > 0 {
			t.Logf("cart not empty")
			if len(items) != countGoroutine {
				t.Errorf("not found items")
			}
		}
	})

	t.Run("repo_delete", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		// init mock
		tracer := mocks.NewTracerMock(t)

		repo := NewInMemoryRepository(tracer)
		defer repo.Close()

		var (
			userID int64 = 4325345
		)

		tracer.StartMock.
			Expect(
				context.Background(),
				"CartRepo:Add",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		for i := 0; i < countGoroutine; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// nolint:gosec
				sku := rand.Int63()
				// nolint:gosec
				count := rand.Uint32()

				err := repo.Add(ctx, model.RequestData{
					UserID: userID,
					Sku:    sku,
					Count:  count,
				})
				if err != nil {
					t.Errorf("Add error: %v", err)
				}
			}()
		}

		wg.Wait()
		tracer.StartMock.
			Expect(
				minimock.AnyContext,
				"CartRepo:GetItemsByUserID",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		items, err := repo.GetItemsByUserID(ctx, model.RequestData{UserID: userID})
		if err != nil {
			t.Errorf("GetItemsByUserID error: %v", err)
		}

		if len(items) < countGoroutine {
			t.Errorf("add item err")
		}
		tracer.StartMock.
			Expect(
				minimock.AnyContext,
				"CartRepo:DeleteItemsBySku",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		for i := 0; i < countGoroutine; i++ {
			num := i
			wg.Add(1)
			go func() {
				defer wg.Done()

				//nolint:govet
				err := repo.DeleteItemsBySku(ctx, model.RequestData{
					UserID: userID,
					Sku:    items[num].SkuID,
				})
				if err != nil {
					t.Logf("DeleteItemsBySku error: %v", err)
				}

			}()
		}

		wg.Wait()
		tracer.StartMock.
			Expect(
				minimock.AnyContext,
				"CartRepo:GetItemsByUserID",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		_, err = repo.GetItemsByUserID(ctx, model.RequestData{UserID: userID})
		if err != nil {
			if !errors.Is(err, model.ErrNotFound) {
				t.Errorf("GetItemsByUserID error: %v", err)
			}
		}
	})

	t.Run("repo_get", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		// init mock
		tracer := mocks.NewTracerMock(t)

		repo := NewInMemoryRepository(tracer)
		defer repo.Close()

		tracer.StartMock.
			Expect(
				context.Background(),
				"CartRepo:Add",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		for i := 0; i < countGoroutine; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				err := repo.Add(ctx, model.RequestData{
					UserID: int64(i),
					Sku:    int64(i),
					//nolint:gosec
					Count: uint32(i),
				})
				if err != nil {
					t.Errorf("Add error: %v", err)
				}
			}()
		}

		wg.Wait()

		tracer.StartMock.
			Expect(
				minimock.AnyContext,
				"CartRepo:GetItemsByUserID",
			).
			Return(minimock.AnyContext, trace.SpanFromContext(context.Background()))

		for i := 0; i < countGoroutine; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				items, err := repo.GetItemsByUserID(ctx, model.RequestData{UserID: int64(i)})
				if err != nil {
					t.Errorf("GetItemsByUserID error: %v", err)
				} else if len(items) > 0 {
					t.Logf("cart not empty")
					//nolint:gosec
					if items[0].SkuID != int64(i) || items[0].Count != uint32(i) {
						t.Error("sku or count not equal expect")
					}
				}
			}()
		}

		wg.Wait()
	})
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
