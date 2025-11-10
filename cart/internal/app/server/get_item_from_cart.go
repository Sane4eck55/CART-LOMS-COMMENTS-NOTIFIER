package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GetItemsByUserID ...
func (s *Server) GetItemsByUserID(w http.ResponseWriter, r *http.Request) {
	data, err := parseRequest(r, int(model.ValidateByUserID))
	if err != nil {
		MakeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	ctx, span := s.tracer.Start(
		r.Context(),
		model.GetItemsByUserIDURL,
		trace.WithAttributes(
			attribute.Int64("UserID", data.UserID),
		),
	)
	defer span.End()

	resp, err := s.cartService.GetItemsFromCart(ctx, *data)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			MakeErrorResponse(w, errors.New(model.ErrNotFound.Error()), http.StatusNotFound)
			return
		}
		MakeErrorResponse(w, err, http.StatusNoContent)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		MakeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
}
