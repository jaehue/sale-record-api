package models

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hublabs/sale-record-api/config"

	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/httpreq"
	"github.com/pangpanglabs/goutils/number"
	"github.com/sirupsen/logrus"
)

type Store struct {
	Id               int64         `json:"id"`
	TenantCode       string        `json:"tenantCode"`
	Code             string        `json:"code"`
	Name             string        `json:"name"`
	TelNo            string        `json:"telNo"`
	AreaCode         string        `json:"areaCode"`
	Area             string        `json:"area"`
	Address          string        `json:"address"`
	AhippingAreaCode string        `json:"shippingAreaCode"`
	AhippingArea     string        `json:"shippingArea"`
	ShippingAddress  string        `json:"shippingAddress"`
	OpenDate         string        `json:"openDate"`
	CloseDate        string        `json:"closeDate"`
	ContractNo       string        `json:"contractNo"`
	StatusCode       string        `json:"statusCode"`
	Enable           bool          `json:"enable"`
	RoundingType     *RoundingType `json:"roundingType"`
	Brands           []StoreBrand  `json:"brands"`
}

type StoreBrand struct {
	Id     int64  `json:"id"`
	Code   string `json:"code"`
	Enable bool   `json:"enable"`
}

type RoundingType struct {
	BaseTrimCode string  `json:"code"`
	Precision    float64 `json:"precision"`
	Offset       int64   `json:"offset"`
}

type StoreList struct {
	Items      []Store `json:"items"`
	TotalCount int64   `json:"totalCount"`
}

func GetStore(ctx context.Context, storeId int64) (*Store, error) {
	var result struct {
		Result  StoreList `json:"result"`
		Success bool      `json:"success"`
		Error   struct {
			Code    int         `json:"code,omitempty"`
			Details interface{} `json:"details,omitempty"`
			Message string      `json:"message,omitempty"`
		} `json:"error"`
	}
	storeUri := fmt.Sprintf("%s/v1/store/getallinfo?storeIds=%d&enable=true&propsEnable=true&withBrand=true&withPayMethod=true&maxResultCount=100&withRoundingType=true",
		config.Config().Services.PlaceManagementAddr, storeId)
	logrus.WithField("storeUri", storeUri).Info("GetStore")
	if _, err := httpreq.New(http.MethodGet, storeUri, nil).
		WithBehaviorLogContext(behaviorlog.FromCtx(ctx)).
		Call(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		logrus.Error("Fail to get store")
		return nil, fmt.Errorf("Get store error:[%d]%s:StoreID=%d", result.Error.Code, result.Error.Message, storeId)
	}
	if result.Result.TotalCount == 0 {
		return nil, fmt.Errorf("Store is null:StoreID=%d", storeId)
	}
	return &result.Result.Items[0], nil
}

func GetRoundSetting(ctx context.Context, storeId int64) (*number.Setting, string, error) {
	store, err := GetStore(ctx, storeId)
	if err != nil {
		return nil, "", err
	}
	if store.RoundingType == nil {
		return nil, "", fmt.Errorf("Store dose not have RoundingType:StoreID=%d", store.Id)
	}
	baseTrimCode := store.RoundingType.BaseTrimCode
	if baseTrimCode == "" || baseTrimCode == "A" {
		return nil, baseTrimCode, nil
	}
	switch baseTrimCode {
	case "C":
		return &number.Setting{
			RoundStrategy: "ceil",
		}, baseTrimCode, nil
	case "O":
		return &number.Setting{
			RoundDigit:    1,
			RoundStrategy: "floor",
		}, baseTrimCode, nil
	case "P":
		return &number.Setting{
			RoundDigit:    1,
			RoundStrategy: "round",
		}, baseTrimCode, nil
	case "Q":
		return &number.Setting{
			RoundDigit:    1,
			RoundStrategy: "ceil",
		}, baseTrimCode, nil
	case "R":
		return &number.Setting{
			RoundStrategy: "round",
		}, baseTrimCode, nil
	case "T":
		return &number.Setting{
			RoundStrategy: "floor",
		}, baseTrimCode, nil
	}
	return nil, "", nil
}
