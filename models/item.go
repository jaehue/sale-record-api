package models

import (
	"context"
	"fmt"
	"net/http"
	"github.com/hublabs/sale-record-api/config"
	"time"

	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/httpreq"
	"github.com/sirupsen/logrus"
)

type Item struct {
	SalePrice float64 `json:"salePrice"`
	Name      string  `json:"name"`
	Offer     *Offer  `json:"offer,omitempty"`
	Sku       *Sku    `json:"sku"`
	BarCode   string  `json:"barCode,,omitempty"`
}

type Offer struct {
	No            string    `json:"no"`
	Name          string    `json:"name"`
	DiscountPrice float64   `json:"discountPrice"`
	StartAt       time.Time `json:"startAt"`
	EndAt         time.Time `json:"endAt"`
}

type Sku struct {
	Product Product `json:"product"`
}
type Product struct {
	Id        int64   `json:"id"`
	Code      string  `json:"code"`
	ListPrice float64 `json:"listPrice"`
	Brand     Brand   `json:"brand"`
}
type Brand struct {
	Id   int64  `json:"id"`
	Code string `json:"code"`
}

func (Item) GetItemByCode(ctx context.Context, code string) (*Item, error) {
	var resp struct {
		Result  Item `json:"result"`
		Success bool `json:"success"`
		Error   struct {
			Code    int    `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"error"`
	}
	url := fmt.Sprintf("%s/v1/items/%s", config.Config().Services.ProductApi, code)
	if _, err := httpreq.New(http.MethodGet, url, nil).
		WithBehaviorLogContext(behaviorlog.FromCtx(ctx)).
		Call(&resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		logrus.Error("Fail to get item")
		return nil, fmt.Errorf("Get item error:[%d]%s", resp.Error.Code, resp.Error.Message)
	}

	return &resp.Result, nil
}

func GetSkuBy(ctx context.Context, skuId int64) (*Sku, error) {
	var resp struct {
		Result  Sku  `json:"result"`
		Success bool `json:"success"`
		Error   struct {
			Code    int    `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"error"`
	}
	url := fmt.Sprintf("%s/v1/skus/%d", config.Config().Services.ProductApi, skuId)
	if _, err := httpreq.New(http.MethodGet, url, nil).
		WithBehaviorLogContext(behaviorlog.FromCtx(ctx)).
		Call(&resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		logrus.Error("Fail to get sku")
		return nil, fmt.Errorf("Get sku error:[%d]%s", resp.Error.Code, resp.Error.Message)
	}

	return &resp.Result, nil
}
