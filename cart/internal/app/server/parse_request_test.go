package server

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ParseRequest(t *testing.T) {
	var (
		// nolint:gosec
		testUserID = rand.Int63()
		// nolint:gosec
		testSku = rand.Int63()
		// nolint:gosec
		testCount = rand.Uint32()
	)

	const testURL = "/user/{user_id}/cart/{sku_id}"
	reader := bytes.NewReader([]byte(fmt.Sprintf(`{"count":%d}`, testCount)))
	req := httptest.NewRequest(http.MethodPost, testURL, reader)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("sku_id", fmt.Sprintf("%d", testSku))
	req.SetPathValue("user_id", fmt.Sprintf("%d", testUserID))

	tests := []struct {
		name      string
		typeValid int

		expectResult model.RequestData
	}{
		{
			name:      "ValidateFull",
			typeValid: int(model.ValidateFull),
			expectResult: model.RequestData{
				UserID: testUserID,
				Sku:    testSku,
				Count:  testCount,
			},
		},
		{
			name:      "ValidateBySku",
			typeValid: int(model.ValidateBySku),
			expectResult: model.RequestData{
				UserID: testUserID,
				Sku:    testSku,
			},
		},
		{
			name:      "ValidateByUserID",
			typeValid: int(model.ValidateByUserID),
			expectResult: model.RequestData{
				UserID: testUserID,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRequest(req, tt.typeValid)
			assert.NoError(t, err)
			assert.Equal(t, &tt.expectResult, result)
		})
	}
}
