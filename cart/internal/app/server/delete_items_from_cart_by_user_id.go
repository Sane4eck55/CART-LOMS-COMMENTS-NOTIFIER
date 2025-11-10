package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DeleteItemsByUserID ...
func (s *Server) DeleteItemsByUserID(w http.ResponseWriter, r *http.Request) {
	data, err := parseRequest(r, int(model.ValidateByUserID))
	if err != nil {
		MakeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	ctx, span := s.tracer.Start(
		r.Context(),
		model.DeleteItemsByUserIDURL,
		trace.WithAttributes(
			attribute.Int64("UserID", data.UserID),
		),
	)
	defer span.End()

	if err := s.cartService.DeleteItemsByUserID(ctx, *data); err != nil {
		if errors.Is(err, model.ErrNoContent) {
			w.Header().Add("Content-Type", "application/json")
			resp := fmt.Errorf("cart emptied successfully")
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
