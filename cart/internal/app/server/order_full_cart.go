// Package server ...
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OrderFullCart ...
func (s *Server) OrderFullCart(w http.ResponseWriter, r *http.Request) {
	data, err := parseRequest(r, int(model.ValidateByUserID))
	if err != nil {
		MakeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	ctx, span := s.tracer.Start(
		r.Context(),
		model.OrderFullCartURL,
		trace.WithAttributes(
			attribute.Int64("UserID", data.UserID),
		),
	)
	defer span.End()

	items, err := s.cartService.GetItemsFromCart(ctx, *data)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			MakeErrorResponse(w, model.ErrCartEmpty, http.StatusNotFound)
			return
		}
		MakeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	orderID, err := s.cartService.OrderCreate(ctx, data.UserID, items)
	if err != nil {
		logger.Errorw(fmt.Sprintf("OrderCreate : %v", err), "span", span)
		MakeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if err = s.cartService.DeleteItemsByUserID(ctx, *data); err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(model.OrderID{OrderID: orderID}); err != nil {
			MakeErrorResponse(w, err, http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(model.OrderID{OrderID: orderID}); err != nil {
		MakeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
}
