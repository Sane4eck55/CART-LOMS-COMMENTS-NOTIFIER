package server

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	var (
		// nolint:gosec
		testUserID = rand.Int63()
		// nolint:gosec
		testSku = rand.Int63()
		// nolint:gosec
		testCount = rand.Uint32()
	)

	t.Run("ValidateUser", func(t *testing.T) {
		t.Parallel()
		data := model.RequestData{
			Sku:   testSku,
			Count: testCount,
		}

		_, err := validate(data, int(model.ValidateFull))
		assert.Error(t, err, errors.New(model.ErrUserIDMoreThanZero))

		_, err = validate(data, int(model.ValidateBySku))
		assert.Error(t, err, errors.New(model.ErrUserIDMoreThanZero))

		_, err = validate(data, int(model.ValidateByUserID))
		assert.Error(t, err, errors.New(model.ErrUserIDMoreThanZero))
	})

	t.Run("ValidateBySku", func(t *testing.T) {
		t.Parallel()
		data := model.RequestData{
			UserID: testUserID,
			Count:  testCount,
		}

		_, err := validate(data, int(model.ValidateFull))
		assert.Error(t, err, errors.New(model.ErrSkuMoreThanZero))

		_, err = validate(data, int(model.ValidateBySku))
		assert.Error(t, err, errors.New(model.ErrSkuMoreThanZero))

	})

	t.Run("ValidateCount", func(t *testing.T) {
		t.Parallel()
		data := model.RequestData{
			UserID: testUserID,
			Sku:    testSku,
		}

		_, err := validate(data, int(model.ValidateFull))
		assert.Error(t, err, errors.New(model.ErrCounItemsMoreThanZero))
	})

}
