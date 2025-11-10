// Package middlewares ...
package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/metrics"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
)

// TimerMiddleware ...
type TimerMiddleware struct {
	h http.Handler
}

// NewTimerMiddleware ...
func NewTimerMiddleware(h http.Handler) http.Handler {
	return &TimerMiddleware{h: h}
}

// ServeHTTP ...
func (m *TimerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	clearURL := getURL(r)
	if clearURL == "" {
		logger.Infow(fmt.Sprintf("default %s, ", r.URL.String()))
		clearURL = "default handler"
	}

	metrics.IncRequestCount(clearURL, model.TypeInternal)

	rr := newResponseWriterWrapper(w)

	m.h.ServeHTTP(rr, r)

	duration := time.Since(start)

	logger.Infow(fmt.Sprintf("%s %s spent %s", r.Method, r.URL.String(), duration))

	metrics.RequestDuration(clearURL, rr.statusCode, model.TypeInternal, duration)
}

func getURL(r *http.Request) string {
	tmp := make(map[string][]string, 3)

	tmp["POST"] = append(tmp["POST"], model.AddItemURL, model.OrderFullCartURL)
	tmp["GET"] = append(tmp["GET"], model.GetItemsByUserIDURL, model.GetMetricsURL)
	tmp["DELETE"] = append(tmp["DELETE"], model.DeleteItemURL, model.DeleteItemsByUserIDURL)

	for method, urls := range tmp {
		if method == r.Method {
			splitInputURL := strings.Split(r.URL.String(), "/")
			for _, url := range urls {
				splitURL := strings.Split(url, " ")
				splitPathURL := strings.Split(splitURL[1], "/")
				if len(splitPathURL) == len(splitInputURL) {
					if compareURL(splitPathURL, splitInputURL) {
						return url
					}
				}
			}
			break
		}
	}
	logger.Infow(fmt.Sprintf("not found url: %s, %s", r.Method, r.URL.String()))
	return ""
}

func compareURL(templURL, inputURL []string) bool {
	if len(templURL) > 1 && len(inputURL) > 1 {
		if templURL[1] != inputURL[1] {
			return false
		}
	}

	if len(templURL) > 3 && len(inputURL) > 3 {
		if templURL[3] != inputURL[3] {
			return false
		}
	}

	return true
}
