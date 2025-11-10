package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
)

func parseRequest(r *http.Request, typeValid int) (*model.RequestData, error) {
	var data model.RequestData

	switch typeValid {
	case int(model.ValidateFull):
		userIDRaw := r.PathValue("user_id")
		userID, _ := strconv.ParseInt(userIDRaw, 10, 64)

		skuRaw := r.PathValue("sku_id")
		sku, _ := strconv.ParseInt(skuRaw, 10, 64)

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return nil, err
		}

		data.Sku = sku
		data.UserID = userID

	case int(model.ValidateBySku):
		userIDRaw := r.PathValue("user_id")
		userID, _ := strconv.ParseInt(userIDRaw, 10, 64)

		skuRaw := r.PathValue("sku_id")
		sku, _ := strconv.ParseInt(skuRaw, 10, 64)

		data.Sku = sku
		data.UserID = userID

	case int(model.ValidateByUserID):
		userIDRaw := r.PathValue("user_id")
		userID, _ := strconv.ParseInt(userIDRaw, 10, 64)

		data.UserID = userID
	}

	validData, err := validate(data, typeValid)
	if err != nil {
		return nil, err
	}

	return validData, nil
}
