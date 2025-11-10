package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DeleteItem ...
func (s *Server) DeleteItem(w http.ResponseWriter, r *http.Request) {
	data, err := parseRequest(r, int(model.ValidateBySku))
	if err != nil {
		MakeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	ctx, span := s.tracer.Start(
		r.Context(),
		model.DeleteItemURL,
		trace.WithAttributes(
			attribute.Int64("UserID", data.UserID),
			attribute.Int64("Sku", data.Sku),
		),
	)
	defer span.End()

	if err := s.cartService.DeleteItem(ctx, *data); err != nil {
		if errors.Is(err, model.ErrNoContent) {
			w.Header().Add("Content-Type", "application/json")
			resp := fmt.Errorf("card %d deleted successfully", data.Sku)
			MakeErrorResponse(w, resp, http.StatusNoContent)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		MakeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
