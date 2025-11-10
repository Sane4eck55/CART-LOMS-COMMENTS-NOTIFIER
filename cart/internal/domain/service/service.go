// Package service ...
package service

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/errgroup"
	pbLoms "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/api/v1"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"
)

// ProductService ...
type ProductService interface {
	GetProductBySku(ctx context.Context, sku int64) (*model.GetProductResponse, error)
}

// Repository ...
type Repository interface {
	Add(ctx context.Context, cartItems model.RequestData) error
	DeleteItemsBySku(ctx context.Context, cartItems model.RequestData) error
	DeleteAllItemsFromCart(ctx context.Context, cartItems model.RequestData) error
	GetItemsByUserID(ctx context.Context, cartItems model.RequestData) ([]model.Cart, error)
	Close()
}

// Loms ...
type Loms interface {
	CreateOrder(ctx context.Context, req *pbLoms.OrderCreateRequest) (*pbLoms.OrderCreateResponse, error)
	GetStocksInfo(ctx context.Context, req *pbLoms.StocksInfoRequest) (*pbLoms.StocksInfoResponse, error)
}

// Tracer ...
type Tracer interface {
	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
}

const (
	// countGoroutines ...
	countGoroutines = 3
)

// Service ...
type Service struct {
	productService ProductService
	Repository     Repository
	loms           Loms
	limiterPS      *rate.Limiter
	tracer         Tracer
}

// NewService ...
func NewService(productService ProductService, repo Repository, loms Loms, limiterPS *rate.Limiter, Tracer Tracer) *Service {
	return &Service{
		productService: productService,
		Repository:     repo,
		loms:           loms,
		limiterPS:      limiterPS,
		tracer:         Tracer,
	}
}

// AddItem ...
func (s *Service) AddItem(ctx context.Context, dataCart model.RequestData) error {
	ctx, span := s.tracer.Start(ctx, "CartService:AddItem")
	defer span.End()
	if err := s.limiterPS.Wait(ctx); err != nil {
		return fmt.Errorf("limiter.Wait: %w", err)
	}

	product, err := s.productService.GetProductBySku(ctx, dataCart.Sku)
	if err != nil {
		return err
	}

	if product.Sku < 1 {
		return model.ErrProductNotFound
	}

	freeStock, err := s.StocksInfo(ctx, product.Sku)
	if err != nil {
		return err
	}

	if dataCart.Count > freeStock {
		return model.ErrAddedMoreItemThanInStock
	}

	if err := s.Repository.Add(ctx, dataCart); err != nil {
		return fmt.Errorf("repository.AddItemsToCart: %w", err)
	}

	return nil
}

// DeleteItem ...
func (s *Service) DeleteItem(ctx context.Context, data model.RequestData) error {
	ctx, span := s.tracer.Start(ctx, "CartService:DeleteItem")
	defer span.End()

	if err := s.Repository.DeleteItemsBySku(ctx, data); err != nil {
		return fmt.Errorf("repository.DeleteItemsBySku: %w", err)
	}

	return nil
}

// DeleteItemsByUserID ...
func (s *Service) DeleteItemsByUserID(ctx context.Context, data model.RequestData) error {
	ctx, span := s.tracer.Start(ctx, "CartService:DeleteItemsByUserID")
	defer span.End()

	if err := s.Repository.DeleteAllItemsFromCart(ctx, data); err != nil {
		return fmt.Errorf("repository.DeleteAllItemsFromCart: %w", err)
	}

	return nil
}

// GetItemsFromCart ...
func (s *Service) GetItemsFromCart(ctx context.Context, data model.RequestData) (*model.GetItemsFromCartResponce, error) {
	ctx, span := s.tracer.Start(ctx, "CartService:GetItemsFromCart")
	defer span.End()

	itemsCart, err := s.Repository.GetItemsByUserID(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("repository.GetItemsByUserID: %w", err)
	}

	if len(itemsCart) < 1 {
		return nil, model.ErrNotFound
	}

	response := &model.GetItemsFromCartResponce{
		Items:      []model.Item{},
		TotalPrice: 0,
	}

	ch := make(chan model.Cart, len(itemsCart))
	chRespItem := make(chan model.Item, len(itemsCart))
	g, gCtx := errgroup.WithContext(ctx)

	for i := 0; i < countGoroutines; i++ {
		g.Go(func() error {
			for item := range ch {
				//nolint:govet
				if err := s.limiterPS.Wait(gCtx); err != nil {
					return fmt.Errorf("limiter.Wait: %w", err)
				}
				//nolint:govet
				itemsPS, err := s.productService.GetProductBySku(gCtx, item.SkuID)
				if err != nil {
					return fmt.Errorf("productService.GetProductBySku: %w", err)
				}

				safePrice, err := SafeInt64ToUint32(itemsPS.Price)
				if err != nil {
					return fmt.Errorf("safeInt64ToUint32: %w", err)
				}

				respItem := model.Item{
					Sku:   item.SkuID,
					Name:  itemsPS.Name,
					Count: item.Count,
					Price: safePrice,
				}

				select {
				case chRespItem <- respItem:
				case <-gCtx.Done():
					return gCtx.Err()
				}
			}
			return nil
		})
	}

	for _, item := range itemsCart {
		ch <- item
	}

	close(ch)

	if err = g.Wait(); err != nil {
		return nil, fmt.Errorf("g.Wait : %v", err)
	}

	close(chRespItem)

	var items []model.Item
	var totalPrice uint32
	for item := range chRespItem {
		items = append(items, item)
		totalPrice += item.Price * item.Count
	}

	sortItem := sortItems(items)
	response.Items = sortItem
	response.TotalPrice = totalPrice

	return response, nil
}

// OrderCreate ...
func (s *Service) OrderCreate(ctx context.Context, UserID int64, items *model.GetItemsFromCartResponce) (int64, error) {
	ctx, span := s.tracer.Start(ctx, "CartService:OrderCreate")
	defer span.End()
	req := convertToOrderCreateRequest(items)
	req.UserID = UserID

	resp, err := s.loms.CreateOrder(ctx, req)
	if err != nil {
		return resp.OrderID, err
	}

	return resp.OrderID, nil
}

// StocksInfo ...
func (s *Service) StocksInfo(ctx context.Context, sku int64) (uint32, error) {
	ctx, span := s.tracer.Start(ctx, "CartService:StocksInfo")
	defer span.End()

	freeStockInfo, err := s.loms.GetStocksInfo(ctx,
		&pbLoms.StocksInfoRequest{
			Sku: sku,
		},
	)
	if err != nil {
		return freeStockInfo.GetCount(), err
	}

	return freeStockInfo.GetCount(), nil
}

func convertToOrderCreateRequest(items *model.GetItemsFromCartResponce) *pbLoms.OrderCreateRequest {
	var req pbLoms.OrderCreateRequest

	for _, item := range items.Items {
		pbItem := &pbLoms.Item{
			Sku:   item.Sku,
			Count: item.Count,
		}

		req.Items = append(req.Items, pbItem)
	}

	return &req
}

// SafeInt64ToUint32 функция преобразования int64 в uint32, чтобы линтер не ругался
func SafeInt64ToUint32(val int64) (uint32, error) {
	if val < 0 || val > int64(math.MaxUint32) {
		return 0, fmt.Errorf("значение %d выходит за границы uint32", val)
	}
	// nolint:gosec
	return uint32(val), nil
}

// sortItems ...
func sortItems(items []model.Item) []model.Item {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Sku < items[j].Sku
	})

	return items
}
