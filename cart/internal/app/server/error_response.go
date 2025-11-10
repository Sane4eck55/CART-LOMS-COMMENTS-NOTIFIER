package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
)

// MakeErrorResponse ...
func MakeErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	type ErrorMessage struct {
		Message string
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResponse := &ErrorMessage{Message: err.Error()}
	if errE := json.NewEncoder(w).Encode(errResponse); errE != nil {
		logger.Infow(fmt.Sprintf("Encode : %v", errE))
		return
	}
}
