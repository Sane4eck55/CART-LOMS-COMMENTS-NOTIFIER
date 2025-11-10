//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type Cart struct {
	suite.Suite
	Host string
}

func TestCart(t *testing.T) {
	t.Parallel()

	suite.RunSuite(t, new(Cart))
}

// BeforeAll выполняется перед запуском тестов
func (c *Cart) BeforeAll(t provider.T) {
	c.Host = "http://localhost:8080"
	t.Logf("host is %v", c.Host)
}

// BeforeEach выполняется перед каждым тестом
func (c *Cart) BeforeEach(t provider.T) {
	t.Feature("Cart")
	t.Tags("Cart", "go")
	t.Owner("Sashka")
}

func (c *Cart) TestCart(t provider.T) {
	t.Title("Добавляем в корзину, получаем данные корзины, удаляем продукт из корзины")

	data := model.RequestData{
		UserID: 1,
		Sku:    1076963,
	}

	t.WithNewStep("Добавляем в корзину", func(t provider.StepCtx) {
		reader := bytes.NewReader([]byte(`{"count":1}`))

		resp, err := http.Post(
			fmt.Sprintf("%v/user/%d/cart/%d",
				c.Host, data.UserID, data.Sku),
			"application/json", reader)
		t.Require().NoError(err, "http post")
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		t.Require().NoError(err, "read all post response body")
		t.Require().Equal(string(respBody), "card add successfully")
		t.Require().Equal(resp.StatusCode, http.StatusOK)
		t.Logf("POST response: %v", string(respBody))
	})

	t.WithNewStep("Получаем данные корзины", func(t provider.StepCtx) {
		resp, err := http.Get(
			fmt.Sprintf("%v/user/%d/cart",
				c.Host, data.UserID),
		)
		t.Require().NoError(err, "http post")
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		items := model.GetItemsFromCartResponce{}
		err = decoder.Decode(&items)
		t.Require().NoError(err, "decode get responce")

		t.Require().Len(items.Items, 1)
		t.Require().Equal(items.Items[0].Sku, data.Sku)
		t.Require().Equal(resp.StatusCode, http.StatusOK)
		t.Logf("items: %v", items)
	})

	t.WithNewStep("Удаляем продукт из корзины", func(t provider.StepCtx) {
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodDelete,
			fmt.Sprintf("%v/user/%d/cart/%d",
				c.Host, data.UserID, data.Sku),
			nil,
		)
		t.Require().NoError(err, "err NewRequest")

		resp, err := client.Do(req)
		t.Require().NoError(err, "err client.Do")
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		t.Require().NoError(err, "read all post response body")
		t.Require().Equal(resp.StatusCode, http.StatusNoContent)
		t.Logf("DELETE response: %v", respBody)
	})

}

func (c *Cart) TestCartCheckout(t provider.T) {
	t.Title("Добавляем в корзину, получаем данные корзины, заказываем продукты из корзины")

	data := model.RequestData{
		UserID: 1,
		Sku:    1076963,
	}

	t.WithNewStep("Добавляем в корзину", func(t provider.StepCtx) {
		reader := bytes.NewReader([]byte(`{"count":1}`))

		resp, err := http.Post(
			fmt.Sprintf("%v/user/%d/cart/%d",
				c.Host, data.UserID, data.Sku),
			"application/json", reader)
		t.Require().NoError(err, "http post")
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		t.Require().NoError(err, "read all post response body")
		t.Require().Equal(string(respBody), "card add successfully")
		t.Require().Equal(resp.StatusCode, http.StatusOK)
		t.Logf("POST response: %v", string(respBody))
	})

	t.WithNewStep("Получаем данные корзины", func(t provider.StepCtx) {
		resp, err := http.Get(
			fmt.Sprintf("%v/user/%d/cart",
				c.Host, data.UserID),
		)
		t.Require().NoError(err, "http post")
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		items := model.GetItemsFromCartResponce{}
		err = decoder.Decode(&items)
		t.Require().NoError(err, "decode get responce")

		t.Require().Len(items.Items, 1)
		t.Require().Equal(items.Items[0].Sku, data.Sku)
		t.Require().Equal(resp.StatusCode, http.StatusOK)
		t.Logf("items: %v", items)
	})

	t.WithNewStep("Оформлячем заказ", func(t provider.StepCtx) {
		resp, err := http.Post(
			fmt.Sprintf("%v/checkout/%d",
				c.Host, data.UserID),
			"application/json", nil)

		t.Require().NoError(err, "http post")
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		orderID := model.OrderID{}
		err = decoder.Decode(&orderID)
		t.Require().NoError(err, "decode get responce")

		t.Require().NotZero(orderID)
		t.Require().Equal(resp.StatusCode, http.StatusOK)

	})

}
