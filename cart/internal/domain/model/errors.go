// Package model ...
package model

import (
	"errors"
)

var (
	// ErrNotOk ...
	ErrNotOk = errors.New("status not ok")
	// ErrNoContent ...
	ErrNoContent = errors.New("no content")
	// ErrNotFound ...
	ErrNotFound = errors.New("not found")
	// ErrManyRequest ...
	ErrManyRequest = errors.New("too many request")

	// ErrProductNotFound product-service
	ErrProductNotFound = errors.New("product not found")
)

var (
	// ErrUserIDMoreThanZero ...
	ErrUserIDMoreThanZero = "Идентификатор пользователя должен быть натуральным числом (больше нуля)"
	// ErrSkuMoreThanZero ...
	ErrSkuMoreThanZero = "SKU должен быть натуральным числом (больше нуля)"
	// ErrCounItemsMoreThanZero ...
	ErrCounItemsMoreThanZero = "Количество должно быть натуральным числом (больше нуля)"
	// ErrSkuNotExists ...
	ErrSkuNotExists = "SKU должен существовать в сервисе product-service"
)

var (
	// ErrCartEmpty ...
	ErrCartEmpty = errors.New("невозможно оформить заказ для пустой корзины")
	// ErrAddedMoreItemThanInStock ...
	ErrAddedMoreItemThanInStock = errors.New("невозможно добавить товара по количеству больше, чем есть в стоках")
)
