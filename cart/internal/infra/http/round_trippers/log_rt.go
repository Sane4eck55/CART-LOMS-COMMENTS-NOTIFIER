// Package roundtrippers ...
package roundtrippers

import (
	"fmt"
	"net/http"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
)

// LogRoundTripper ...
type LogRoundTripper struct {
	rt http.RoundTripper
}

// NewLogRoundTripper ...
func NewLogRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &LogRoundTripper{rt: rt}
}

// RoundTrip ...
func (l *LogRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	logger.Infow(fmt.Sprintf("%s called\n", r.URL.String()))

	resp, err := l.rt.RoundTrip(r)
	logger.Infow(fmt.Sprintf("%+v, %v", resp, err))
	return resp, err
}
