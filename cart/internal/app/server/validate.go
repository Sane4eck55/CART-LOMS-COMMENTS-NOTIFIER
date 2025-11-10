package server

import (
	"errors"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/go-playground/validator/v10"
)

// валидация входных значений
func validate(data model.RequestData, typeValid int) (*model.RequestData, error) {
	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "UserID" {
				return nil, errors.New(model.ErrUserIDMoreThanZero)
			} else if err.Field() == "Sku" && (typeValid == int(model.ValidateFull) || typeValid == int(model.ValidateBySku)) {
				return nil, errors.New(model.ErrSkuMoreThanZero)
			} else if err.Field() == "Count" && typeValid == int(model.ValidateFull) {
				return nil, errors.New(model.ErrCounItemsMoreThanZero)
			}
		}
	}

	return &data, nil
}
