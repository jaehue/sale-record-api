package models

import (
	"time"
)

type SaleRecordSearchInput struct {
	CustomerId int64 `query:"customerId"`
	// TransactionStartAt time.Time `json:"-"`
	// TransactionEndAt   time.Time `json:"-"`
	OrderStartAt      string    `query:"orderStartAt"`
	OrderEndAt        string    `query:"orderEndAt"`
	OrderStartTime    time.Time `query:"-"`
	OrderEndTime      time.Time `query:"-"`
	CreatedId         int64     `query:"createdId"`
	TransactionStatus string    `query:"status"`
	TransactionType   string    `query:"type"`
	ChannelType       string    `query:"channelType"`
	SalesmanId        int64     `query:"salesmanId"`
	EmpId             string    `query:"empId"`
	StoreId           int64     `query:"storeId"`
	TransactionId     int64     `query:"transactionId"`
	OrderId           int64     `query:"orderId"`
	RefundId          int64     `query:"refundId"`
	OrderIds          string    `query:"orderIds"`
	RefundIds         string    `query:"refundIds"`
	OuterOrderNo      string    `query:"outerOrderNo"`
	IsOutPaid         string    `query:"isOutPaid"`
	SkipCount         int       `query:"skipCount"`
	MaxResultCount    int       `query:"maxResultCount"`
}
type SaleRecordLogSearchInput struct {
	StartAt         string    `query:"startAt"`
	EndAt           string    `query:"endAt"`
	StartTime       time.Time `query:"-"`
	EndTime         time.Time `query:"-"`
	ErrType         string    `query:"errType"`
	ChannelType     string    `query:"channelType"`
	TransactionType string    `query:"transactionType"`
	IsSuccess       bool      `query:"isSuccess"`
	StoreId         int64     `query:"storeId"`
	OrderId         int64     `query:"orderId"`
	RefundId        int64     `query:"refundId"`
	SkipCount       int       `query:"skipCount"`
	MaxResultCount  int       `query:"maxResultCount"`
}
type SearchSaleRecordOutput struct {
	TotalCount int64                `json:"totalCount"`
	Items      []AssortedSaleRecord `json:"items"`
}

type SearchSaleRecordLogOutPut struct {
	TotalCount int64               `json:"totalCount"`
	Items      []SaleRecordSuccess `json:"items"`
}
