// Package retryclient ...
package retryclient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
)

// HTTPClientConfig ...
type HTTPClientConfig struct {
	Client     *http.Client
	MaxRetries int
	Delay      time.Duration
}

// NewRetryClient ...
func NewRetryClient(client *http.Client, maxRetries int, delay time.Duration) *HTTPClientConfig {
	return &HTTPClientConfig{
		Client:     client,
		MaxRetries: maxRetries,
		Delay:      delay,
	}
}

// RetryMiddleware ...
func (c *HTTPClientConfig) RetryMiddleware() func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		var err error

		for attempt := 1; attempt <= c.MaxRetries; attempt++ {
			resp, err = c.Client.Do(req)
			if err != nil {
				return nil, err
			}

			// если не 420 или 429, то возвращаем резлуьтат запроса
			if resp.StatusCode != 420 && resp.StatusCode != 429 {
				return resp, nil
			}

			// если статус 420 или 429 делаем ретраи
			// проверяем кол-во ретраев, если меньше, то засыпаем , в противном случае позвращаем ошибку
			if attempt < c.MaxRetries {
				logger.Errorw(fmt.Sprintf("Attempt %d failed with status %d. Retrying...\n", attempt, resp.StatusCode))
				time.Sleep(c.Delay)
			} else {
				return resp, model.ErrManyRequest
			}
		}
		return resp, nil
	}
}
