// Package productservice ...
package productservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	retryclient "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/http/retry_client"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/metrics"
)

const (
	// ProductServiceTimeOut ...
	ProductServiceTimeOut = 10 * time.Second
)

// ErrNotOk ...
var ErrNotOk = errors.New("status not ok")

// ProductService ...
type ProductService struct {
	httpClient retryclient.HTTPClientConfig
	token      string
	address    string
}

// NewProductService ...
func NewProductService(httpClient retryclient.HTTPClientConfig, token string, address string) *ProductService {
	return &ProductService{
		httpClient: httpClient,
		token:      token,
		address:    address,
	}
}

// GetProductBySku ...
func (ps *ProductService) GetProductBySku(ctx context.Context, sku int64) (*model.GetProductResponse, error) {
	var (
		response   *http.Response
		statusCode int
	)

	ctx, cancel := context.WithTimeout(ctx, ProductServiceTimeOut)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/product/%d", ps.address, sku),
		http.NoBody,
	)
	if err != nil {
		return nil, model.ErrNoContent
	}

	req.Header.Add("X-API-KEY", ps.token)

	start := time.Now()
	defer func() {
		metrics.RequestDuration(model.GetProductBySkuURL, statusCode, model.TypeExternal, time.Since(start))
	}()

	doRequest := ps.httpClient.RetryMiddleware()
	metrics.IncRequestCount(model.GetProductBySkuURL, model.TypeExternal)
	response, err = doRequest(req)
	if err != nil {
		if errors.Is(err, model.ErrManyRequest) {
			statusCode = http.StatusTooManyRequests
			return nil, model.ErrManyRequest
		}
		statusCode = http.StatusForbidden
		return nil, err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			return
		}

	}()

	if response != nil {
		statusCode = response.StatusCode
	}

	if response.StatusCode == http.StatusNotFound {
		return nil, model.ErrProductNotFound
	}

	if response.StatusCode != http.StatusOK {
		return nil, ErrNotOk
	}

	resp := &model.GetProductResponse{}
	if err := json.NewDecoder(response.Body).Decode(resp); err != nil {
		return nil, fmt.Errorf("json.NewDecoder: %w", err)
	}

	return resp, nil
}
