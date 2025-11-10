//go:build e2e

package repository

import (
	"context"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/repository/postgres/connect"
	repo "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/repository/sqlc"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LomsRepo struct {
	suite.Suite
	masterDsn  string
	replicaDsn string
	repo       *repo.Repo
	infoStocks map[int64][]int64
}

func TestLoms(t *testing.T) {
	t.Parallel()

	suite.RunSuite(t, new(LomsRepo))
}

// BeforeAll выполняется перед запуском тестов
func (r *LomsRepo) BeforeAll(t provider.T) {
	r.masterDsn = "postgresql://user:password@localhost:5432/route256?sslmode=disable"
	t.Logf("master dsn is %v", r.masterDsn)
	r.replicaDsn = "postgresql://user:password@localhost:5433/route256?sslmode=disable"
	t.Logf("replica dsn is %v", r.replicaDsn)

	ctx := context.Background()

	masterPool, err := connect.NewPool(ctx, r.masterDsn)
	require.NoError(t, err)
	replicaPool, err := connect.NewPool(ctx, r.replicaDsn)
	require.NoError(t, err)
	r.repo = repo.NewRepo(masterPool, replicaPool)

	testSkus := []int64{111111, 22222, 33333}
	tmpMap := map[int64][]int64{}

	for _, sku := range testSkus {
		stocks, _ := r.repo.GetStocksBySku(ctx, sku)
		tmpMap[stocks.Sku] = append(tmpMap[stocks.Sku], int64(stocks.TotalCount), int64(stocks.Reserved))
	}

	r.infoStocks = tmpMap
}

func (r *LomsRepo) TestLomsRepo_CreateOrder(t provider.T) {
	t.Title("Создание заказа")
	ctx := context.Background()

	usersOrders := model.Order{
		UserID: 1,
		Items: []model.Item{
			{
				Sku:   111111,
				Count: 2,
			},
			{
				Sku:   22222,
				Count: 56,
			},
		},
	}

	orderID, err := r.repo.CreateOrder(ctx, usersOrders)
	require.NoError(t, err)
	assert.NotZero(t, orderID)

	t.Cleanup(func() {
		r.repo.Delete(ctx, orderID)
		for _, item := range usersOrders.Items {
			r.repo.UpdateStocks(ctx, item.Sku, r.infoStocks[item.Sku][0], r.infoStocks[item.Sku][1])
		}
	})
}

func (r *LomsRepo) TestLomsRepo_SetStatusOrder(t provider.T) {
	t.Title("Смена статуса заказа")
	ctx := context.Background()

	usersOrders := model.Order{
		UserID: 55,
		Items: []model.Item{
			{
				Sku:   111111,
				Count: 2,
			},
			{
				Sku:   22222,
				Count: 56,
			},
		},
	}

	orderID, err := r.repo.CreateOrder(ctx, usersOrders)
	require.NoError(t, err)

	info, err := r.repo.GetInfoByOrderIDMaster(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, info.Status, model.StatusOrderNew)

	err = r.repo.SetStatusOrder(ctx, orderID, model.StatusOrderAwaitingPayment)
	require.NoError(t, err)

	info, err = r.repo.GetInfoByOrderIDMaster(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, info.Status, model.StatusOrderAwaitingPayment)

	t.Cleanup(func() {
		r.repo.Delete(ctx, orderID)
		for _, item := range usersOrders.Items {
			r.repo.UpdateStocks(ctx, item.Sku, r.infoStocks[item.Sku][0], r.infoStocks[item.Sku][1])
		}
	})
}

func (r *LomsRepo) TestLomsRepo_GetInfoByOrderID(t provider.T) {
	t.Title("Получит информацию о заказе")
	ctx := context.Background()

	usersOrders := model.Order{
		UserID: 55,
		Items: []model.Item{
			{
				Sku:   22222,
				Count: 4,
			},
			{
				Sku:   33333,
				Count: 9,
			},
		},
	}

	orderID, err := r.repo.CreateOrder(ctx, usersOrders)
	require.NoError(t, err)

	info, err := r.repo.GetInfoByOrderIDMaster(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, info.UserID, usersOrders.UserID)
	require.Equal(t, info.Status, model.StatusOrderNew)
	require.Equal(t, info.Items[0].Sku, usersOrders.Items[0].Sku)
	require.Equal(t, info.Items[0].Count, usersOrders.Items[0].Count)
	require.Equal(t, info.Items[1].Sku, usersOrders.Items[1].Sku)
	require.Equal(t, info.Items[1].Count, usersOrders.Items[1].Count)

	t.Cleanup(func() {
		r.repo.Delete(ctx, orderID)
		for _, item := range usersOrders.Items {
			r.repo.UpdateStocks(ctx, item.Sku, r.infoStocks[item.Sku][0], r.infoStocks[item.Sku][1])
		}
	})
}

func (r *LomsRepo) TestLomsRepo_GetFreeStocksBySku(t provider.T) {
	t.Title("Получение информарции о стоке на ску")
	ctx := context.Background()

	var (
		sku         int64  = 111111
		expectCount uint32 = 266
	)

	freeStock, err := r.repo.GetFreeStocksBySkuMaster(ctx, sku)
	require.NoError(t, err)
	t.Log("freeStock :", freeStock)
	require.Equal(t, freeStock, expectCount)
}

func (r *LomsRepo) TestLomsRepo_Reserver(t provider.T) {
	t.Title("Резерв стока")
	ctx := context.Background()

	items := []model.Item{
		{
			Sku:   111111,
			Count: 55,
		},
		{
			Sku:   33333,
			Count: 14,
		},
	}

	freeStockSku1, err := r.repo.GetFreeStocksBySkuMaster(ctx, items[0].Sku)
	require.NoError(t, err)

	freeStockSku2, err := r.repo.GetFreeStocksBySkuMaster(ctx, items[1].Sku)
	require.NoError(t, err)

	err = r.repo.Reserve(ctx, items)
	require.NoError(t, err)

	newFreeStockSku1, err := r.repo.GetFreeStocksBySkuMaster(ctx, items[0].Sku)
	require.NoError(t, err)

	newFreeStockSku2, err := r.repo.GetFreeStocksBySkuMaster(ctx, items[1].Sku)
	require.NoError(t, err)

	require.Equal(t, newFreeStockSku1, freeStockSku1-items[0].Count)
	require.Equal(t, newFreeStockSku2, freeStockSku2-items[1].Count)
}

func (r *LomsRepo) TestLomsRepo_ReserveRemove(t provider.T) {
	t.Title("Удаление резерва стока(товар купили)")
	ctx := context.Background()

	items := []model.Item{
		{
			Sku:   33333,
			Count: 13,
		},
	}

	freeStock, err := r.repo.GetFreeStocksBySkuMaster(ctx, items[0].Sku)
	require.NoError(t, err)

	err = r.repo.Reserve(ctx, items)
	require.NoError(t, err)

	freeStockAfterReserve, err := r.repo.GetFreeStocksBySkuMaster(ctx, items[0].Sku)
	require.NoError(t, err)
	require.Equal(t, freeStockAfterReserve, freeStock-items[0].Count)

	err = r.repo.ReserveRemove(ctx, items[0])
	require.NoError(t, err)

	newFreeStockSku1, err := r.repo.GetFreeStocksBySkuMaster(ctx, items[0].Sku)
	require.NoError(t, err)
	require.Equal(t, newFreeStockSku1, freeStock-items[0].Count)
}

func (r *LomsRepo) TestLomsRepo_ReserveCancel(t provider.T) {
	t.Title("Отмена резерва стока")
	ctx := context.Background()

	items := []model.Item{
		{
			Sku:   111111,
			Count: 55,
		},
	}

	freeStock, err := r.repo.GetFreeStocksBySkuMaster(ctx, items[0].Sku)
	require.NoError(t, err)

	err = r.repo.Reserve(ctx, items)
	require.NoError(t, err)

	freeStockAfterReserve, err := r.repo.GetFreeStocksBySkuMaster(ctx, items[0].Sku)
	require.NoError(t, err)
	require.Equal(t, freeStockAfterReserve, freeStock-items[0].Count)

	err = r.repo.ReserveCancel(ctx, items[0])
	require.NoError(t, err)

	newFreeStockSku1, err := r.repo.GetFreeStocksBySkuMaster(ctx, items[0].Sku)
	require.NoError(t, err)
	require.Equal(t, newFreeStockSku1, freeStock)
}
