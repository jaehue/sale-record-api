package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hublabs/sale-record-api/config"

	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/httpreq"
	"github.com/sirupsen/logrus"
)

type CustomerInfo struct {
	Id               int64   `json:"id"`
	HrEmpNo          string  `json:"hrEmpNo"`
	CardNo           string  `json:"cardNo"` /*会员卡号*/
	TenantCode       string  `json:"tenantCode"`
	MallId           int64   `json:"mallId"`
	Mobile           string  `json:"mobile"`
	CardTypeId       int64   `json:"cardTypeId"`
	GradeId          int64   `json:"gradeId"`
	Birthday         string  `json:"birthday"`
	CurrentPoints    float64 `json:"currentPoints"`
	TotalSaleAmount  float64 `json:"totalSaleAmount"`
	TotalSaleCount   int64   `json:"totalSaleCount"`
	Status           string  `json:"status"` /*是否有效*/
	JoinDate         string  `json:"joinDate"`
	ConsumePoints    float64 `json:"consumePoints"`
	ExchangeAmount   float64 `json:"exchangeAmount"`
	MinConsumePoints float64 `json:"minConsumePoints"`
	MaxConsumeRate   int64   `json:"maxConsumeRate"`
	ConsumeUnit      int64   `json:"consumeUnit"`
	GradeName        string  `json:"gradeName"`
	MemberName       string  `json:"memberName"`
}

type MileageMstResult struct {
	Items []MileageMst `json:"items"`
}

type MileageMst struct {
	PointPrice  float64      `json:"pointPrice"`
	MileageDtls []MileageDtl `json:"mileageDtls"`
}

type MileageDtl struct {
	ItemId     int64   `json:"itemId"`
	PointPrice float64 `json:"pointPrice"`
}

type Coupon struct {
	CouponNo   string `json:"couponNo"`
	IsInternal bool   `json:"isInternal"`
}

func (MileageMst) getUsedMileageMst(ctx context.Context, orderId int64, tenantCode string) (*MileageMst, error) {
	var resp struct {
		Result  *MileageMstResult `json:"result"`
		Success bool              `json:"success"`
		Error   struct {
			Code    int    `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"error"`
	}
	url := fmt.Sprintf("%s/v1/mileage?tenantCode=%s&tradeNo=%d&type=M", config.Config().Services.MembershipBenefitAddr, tenantCode, orderId)
	if _, err := httpreq.New(http.MethodGet, url, nil).
		WithBehaviorLogContext(behaviorlog.FromCtx(ctx)).
		Call(&resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		logrus.Error("Fail to get userdMileage")
		return nil, fmt.Errorf("Get usedMileage error:[%d]%s", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result != nil && len(resp.Result.Items) > 0 {
		return &resp.Result.Items[0], nil
	}
	return nil, nil
}

func (mileageMst *MileageMst) getUsedMileagePrice(itemId int64) float64 {
	if mileageMst == nil {
		return 0
	}

	for _, mileageDtl := range mileageMst.MileageDtls {
		if mileageDtl.ItemId != itemId {
			continue
		}
		return mileageDtl.PointPrice
	}
	return 0
}

func GetCoupon(ctx context.Context, couponNo string) (*Coupon, error) {
	var resp struct {
		Result  Coupon `json:"result"`
		Success bool   `json:"success"`
		Error   struct {
			Code    int    `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"error"`
	}
	url := fmt.Sprintf("%s/v1/coupons/%s", config.Config().Services.CouponAddr, couponNo)
	if _, err := httpreq.New(http.MethodGet, url, nil).
		WithBehaviorLogContext(behaviorlog.FromCtx(ctx)).
		Call(&resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		logrus.Error("Fail to get Coupon")
		return nil, fmt.Errorf("Get Coupon error:[%d]%s", resp.Error.Code, resp.Error.Message)
	}
	return &resp.Result, nil
}

func GetMember(ctx context.Context, customerId int64, tenantCode string) (*CustomerInfo, error) {
	var (
		result struct {
			Result  CustomerInfo `json:"result"`
			Success bool         `json:"success"`
			Error   struct{}     `json:"error"`
		}
	)
	memberUri := fmt.Sprintf("%s/v1/member?memberId=%d&tenantCode=%s", config.Config().Services.MembershipAddr, customerId, tenantCode)
	if _, err := httpreq.New(http.MethodGet, memberUri, nil).
		WithBehaviorLogContext(behaviorlog.FromCtx(ctx)).
		Call(&result); err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, errors.New("Fail to get customer")
	}

	return &result.Result, nil
}
