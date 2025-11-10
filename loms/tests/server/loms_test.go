//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/pkg/api/v1"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type Loms struct {
	suite.Suite
	Host string
}

func TestLoms(t *testing.T) {
	t.Parallel()

	suite.RunSuite(t, new(Loms))
}

// BeforeAll выполняется перед запуском тестов
func (c *Loms) BeforeAll(t provider.T) {
	c.Host = "http://localhost:8084"
	t.Logf("host is %v", c.Host)
}

// BeforeEach выполняется перед каждым тестом
func (c *Loms) BeforeEach(t provider.T) {
	t.Feature("Loms")
	t.Tags("Loms", "go")
	t.Owner("Sashka")
}

func (c *Loms) TestLoms_Pay(t provider.T) {
	t.Title("Создание заказа, инфомрация о заказе, оплата закза, информация о стоке")
	orderID := pb.OrderCreateResponse{}

	reqCreate := &pb.OrderCreateRequest{
		UserID: 1,
		Items: []*pb.Item{
			{
				Sku:   1076963,
				Count: 10,
			},
			{
				Sku:   135937324,
				Count: 50,
			},
		},
	}
	var stockSku1, stockSku2 uint32
	stocksBySku := []uint32{stockSku1, stockSku2}

	t.WithNewStep("Инфомрация о стоке Sku", func(t provider.StepCtx) {
		for i, item := range reqCreate.Items {
			resp, err := http.Get(
				fmt.Sprintf("%v/stock/info?sku=%d",
					c.Host, item.Sku),
			)
			t.Require().NoError(err, "http get")
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			t.Require().NoError(err, "не удалось считать ответ")
			t.WithNewAttachment("response body", allure.JSON, respBody)

			stockInfo := &pb.StocksInfoResponse{}
			err = json.Unmarshal(respBody, stockInfo)
			t.Require().NoError(err, "парсинг ответа")

			stocksBySku[i] = stockInfo.Count

			t.Require().Equal(resp.StatusCode, http.StatusOK)
			t.Logf("Get response: %v", string(respBody))
		}
	})

	t.WithNewStep("Добавляем в корзину", func(t provider.StepCtx) {
		jsonData, err := json.Marshal(reqCreate)
		if err != nil {
			logger.Infow("Error marshalling:", err)
			return
		}
		t.WithNewAttachment("request body", allure.JSON, jsonData)
		reader := bytes.NewReader([]byte(jsonData))

		resp, err := http.Post(
			fmt.Sprintf("%v/order/create",
				c.Host),
			"application/json", reader)
		t.Require().NoError(err, "http post")
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body) //{"OrderID":"1"}
		t.Require().NoError(err, "не удалось считать ответ")
		t.WithNewAttachment("response body", allure.JSON, respBody)

		strOrderID := model.OrderIDTest{}
		err = json.Unmarshal(respBody, &strOrderID)
		t.Require().NoError(err, "парсинг ответа")

		tmpID, _ := strconv.ParseInt(strOrderID.OrderID, 10, 64)
		orderID.OrderID = int64(tmpID)

		t.Require().NotZero(orderID.OrderID)
		t.Require().Equal(resp.StatusCode, http.StatusOK)
		t.Logf("POST response: %v", respBody)
	})

	t.WithNewStep("Инфомрация о заказе", func(t provider.StepCtx) {
		resp, err := http.Get(
			fmt.Sprintf("%v/order/info?orderId=%d",
				c.Host, orderID.OrderID),
		)
		t.Require().NoError(err, "http get")
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		t.Require().NoError(err, "не удалось считать ответ")
		t.WithNewAttachment("response body", allure.JSON, respBody)

		orderInfo := &model.OrderCreateResponseTest{}
		err = json.Unmarshal(respBody, orderInfo)
		t.Require().NoError(err, "парсинг ответа")

		tmpSku1, _ := strconv.ParseInt(orderInfo.Items[0].Sku, 10, 64)
		tmpSku2, _ := strconv.ParseInt(orderInfo.Items[1].Sku, 10, 64)

		t.Require().Len(orderInfo.Items, 2)
		t.Require().Equal(tmpSku1, reqCreate.Items[0].Sku)
		t.Require().Equal(orderInfo.Items[0].Count, reqCreate.Items[0].Count)
		t.Require().Equal(tmpSku2, reqCreate.Items[1].Sku)
		t.Require().Equal(orderInfo.Items[1].Count, reqCreate.Items[1].Count)
		t.Require().Equal(resp.StatusCode, http.StatusOK)
		t.Logf("GET response: %v", string(respBody))
	})

	t.WithNewStep("Оплата заказа", func(t provider.StepCtx) {
		jsonData, err := json.Marshal(&pb.OrderPayRequest{
			OrderID: orderID.OrderID,
		})
		if err != nil {
			logger.Infow("Error marshalling:", err)
			return
		}

		reader := bytes.NewReader([]byte(jsonData))

		resp, err := http.Post(
			fmt.Sprintf("%v/order/pay",
				c.Host),
			"application/json", reader)
		t.Require().NoError(err, "http post")
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		t.Require().NoError(err, "не удалось считать ответ")
		t.WithNewAttachment("response body", allure.JSON, respBody)

		t.Log("respBody : ", string(respBody))

		t.Require().Equal(resp.StatusCode, http.StatusOK)
		t.Logf("POST response: %v", string(respBody))

	})

	t.WithNewStep("Инфомрация о стоке", func(t provider.StepCtx) {
		for i, item := range reqCreate.Items {
			resp, err := http.Get(
				fmt.Sprintf("%v/stock/info?sku=%d",
					c.Host, item.Sku),
			)
			t.Require().NoError(err, "http get")
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			t.Require().NoError(err, "не удалось считать ответ")
			t.WithNewAttachment("response body", allure.JSON, respBody)

			stockInfo := &pb.StocksInfoResponse{}
			err = json.Unmarshal(respBody, stockInfo)
			t.Require().NoError(err, "парсинг ответа")

			t.Require().Equal(stockInfo.Count, stocksBySku[i]-item.Count)

			t.Require().Equal(resp.StatusCode, http.StatusOK)
			t.Logf("Get response: %v", string(respBody))
		}
	})
}

func (c *Loms) TestLoms_Cancel(t provider.T) {
	t.Title("Создание заказа, инфомрация о заказе, отмена закза, информация о стоке")
	orderID := pb.OrderCreateResponse{}

	reqCreate := &pb.OrderCreateRequest{
		UserID: 1,
		Items: []*pb.Item{
			{
				Sku:   1076963,
				Count: 10,
			},
			{
				Sku:   135937324,
				Count: 50,
			},
		},
	}
	var stockSku1, stockSku2 uint32
	stocksBySku := []uint32{stockSku1, stockSku2}

	t.WithNewStep("Инфомрация о стоке Sku", func(t provider.StepCtx) {
		for i, item := range reqCreate.Items {
			resp, err := http.Get(
				fmt.Sprintf("%v/stock/info?sku=%d",
					c.Host, item.Sku),
			)
			t.Require().NoError(err, "http get")
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			t.Require().NoError(err, "не удалось считать ответ")
			t.WithNewAttachment("response body", allure.JSON, respBody)

			stockInfo := &pb.StocksInfoResponse{}
			err = json.Unmarshal(respBody, stockInfo)
			t.Require().NoError(err, "парсинг ответа")

			stocksBySku[i] = stockInfo.Count

			t.Require().Equal(resp.StatusCode, http.StatusOK)
			t.Logf("Get response: %v", string(respBody))
		}
	})

	t.WithNewStep("Добавляем в корзину", func(t provider.StepCtx) {
		jsonData, err := json.Marshal(reqCreate)
		if err != nil {
			logger.Infow("Error marshalling:", err)
			return
		}
		t.WithNewAttachment("request body", allure.JSON, jsonData)
		reader := bytes.NewReader([]byte(jsonData))

		resp, err := http.Post(
			fmt.Sprintf("%v/order/create",
				c.Host),
			"application/json", reader)
		t.Require().NoError(err, "http post")
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body) //{"OrderID":"1"}
		t.Require().NoError(err, "не удалось считать ответ")
		t.WithNewAttachment("response body", allure.JSON, respBody)
		t.Log("respBody :", string(respBody))
		strOrderID := model.OrderIDTest{}
		err = json.Unmarshal(respBody, &strOrderID)
		t.Require().NoError(err, "парсинг ответа")

		tmpID, _ := strconv.ParseInt(strOrderID.OrderID, 10, 64)
		orderID.OrderID = int64(tmpID)

		t.Require().NotZero(orderID.OrderID)
		t.Require().Equal(resp.StatusCode, http.StatusOK)
		t.Logf("POST response: %v", respBody)
	})

	t.WithNewStep("Инфомрация о заказе", func(t provider.StepCtx) {
		resp, err := http.Get(
			fmt.Sprintf("%v/order/info?orderId=%d",
				c.Host, orderID.OrderID),
		)
		t.Require().NoError(err, "http get")
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		t.Log("respBody : ", respBody)
		t.Require().NoError(err, "не удалось считать ответ")
		t.WithNewAttachment("response body", allure.JSON, respBody)

		orderInfo := &model.OrderCreateResponseTest{}
		err = json.Unmarshal(respBody, orderInfo)
		t.Require().NoError(err, "парсинг ответа")

		tmpSku1, _ := strconv.ParseInt(orderInfo.Items[0].Sku, 10, 64)
		tmpSku2, _ := strconv.ParseInt(orderInfo.Items[1].Sku, 10, 64)

		t.Require().Len(orderInfo.Items, 2)
		t.Require().Equal(tmpSku1, reqCreate.Items[0].Sku)
		t.Require().Equal(orderInfo.Items[0].Count, reqCreate.Items[0].Count)
		t.Require().Equal(tmpSku2, reqCreate.Items[1].Sku)
		t.Require().Equal(orderInfo.Items[1].Count, reqCreate.Items[1].Count)
		t.Require().Equal(resp.StatusCode, http.StatusOK)
		t.Logf("POST response: %v", string(respBody))
	})

	t.WithNewStep("Отмена заказа", func(t provider.StepCtx) {
		jsonData, err := json.Marshal(&pb.OrderCancelRequest{
			OrderID: orderID.OrderID,
		})
		if err != nil {
			logger.Infow("Error marshalling:", err)
			return
		}

		reader := bytes.NewReader([]byte(jsonData))

		resp, err := http.Post(
			fmt.Sprintf("%v/order/cancel",
				c.Host),
			"application/json", reader)
		t.Require().NoError(err, "http post")
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		t.Require().NoError(err, "не удалось считать ответ")
		t.WithNewAttachment("response body", allure.JSON, respBody)

		t.Log("respBody : ", string(respBody))

		t.Require().Equal(resp.StatusCode, http.StatusOK)
		t.Logf("POST response: %v", string(respBody))

	})

	t.WithNewStep("Инфомрация о стоке", func(t provider.StepCtx) {
		for i, item := range reqCreate.Items {
			resp, err := http.Get(
				fmt.Sprintf("%v/stock/info?sku=%d",
					c.Host, item.Sku),
			)
			t.Require().NoError(err, "http get")
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			t.Require().NoError(err, "не удалось считать ответ")
			t.WithNewAttachment("response body", allure.JSON, respBody)

			stockInfo := &pb.StocksInfoResponse{}
			err = json.Unmarshal(respBody, stockInfo)
			t.Require().NoError(err, "парсинг ответа")

			t.Require().Equal(stockInfo.Count, stocksBySku[i])

			t.Require().Equal(resp.StatusCode, http.StatusOK)
			t.Logf("Get response: %v", string(respBody))
		}
	})
}
