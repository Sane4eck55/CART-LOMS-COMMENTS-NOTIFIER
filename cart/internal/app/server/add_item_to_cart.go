// Package server ...
package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AddItem ...
func (s *Server) AddItem(w http.ResponseWriter, r *http.Request) {
	data, err := parseRequest(r, int(model.ValidateFull))
	if err != nil {
		MakeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	ctx, span := s.tracer.Start(
		r.Context(),
		model.AddItemURL,
		trace.WithAttributes(
			attribute.Int64("UserID", data.UserID),
			attribute.Int64("Sku", data.Sku),
			attribute.Int64("Count", int64(data.Count)),
		),
	)
	defer span.End()

	if err = s.cartService.AddItem(ctx, *data); err != nil {
		if errors.Is(err, model.ErrProductNotFound) {
			MakeErrorResponse(w, errors.New(model.ErrSkuNotExists), http.StatusPreconditionFailed)
			return
		}
		if errors.Is(err, model.ErrManyRequest) {
			MakeErrorResponse(w, errors.New(model.ErrSkuNotExists), http.StatusTooManyRequests)
			return
		}
		if errors.Is(err, model.ErrAddedMoreItemThanInStock) {
			MakeErrorResponse(w, err, http.StatusPreconditionFailed)
			return
		}
		MakeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write([]byte("card add successfully"))
	if err != nil {
		logger.Infow(fmt.Sprintf("err w.Write : %v", err))
		return
	}

}
