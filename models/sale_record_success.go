package models

import (
	"context"
	"encoding/json"
	"github.com/hublabs/sale-record-api/factory"
	"time"

	"github.com/go-xorm/xorm"
)

type ErrorType string

const (
	SaleRecordError                   ErrorType = "SaleRecord"
	CreatedIdError                    ErrorType = "CreatedId"
	POSSalesmanIdError                ErrorType = "POSSalesmanId"
	StoreIdError                      ErrorType = "StoreId"
	ItemFeeRateError                  ErrorType = "ItemFeeRate"
	TotalPriceError                   ErrorType = "TotalPrice"
	DiscountPriceError                ErrorType = "DiscountPrice"
	DiscountPriceNotMatchOfferError   ErrorType = "DiscountPriceNotMatchOffer"
	OfferNoError                      ErrorType = "OfferNo"
	StoreNotExistError                ErrorType = "StoreNotExist"
	BrandNotMatchError                ErrorType = "BrandNotMatch"
	SkuNotExistError                  ErrorType = "SkuNotExist"
	SkuListPriceError                 ErrorType = "SkuListPrice"
	ProductNotMatchError              ErrorType = "ProductNotMatch"
	ProductBrandError                 ErrorType = "ProductBrand"
	DistributedCashPriceError         ErrorType = "DistributedCashPrice"
	TotalDistributedPaymentPriceError ErrorType = "TotalDistributedPaymentPrice"
	MileageError                      ErrorType = "Mileage"
	PayMentNoExistError               ErrorType = "PayMentNotExist"
)

type SaleRecordSuccess struct {
	Id                    int64     `json:"id"`
	TenantCode            string    `json:"tenantCode" xorm:"index VARCHAR(50) notnull" validate:"required"`
	ChannelType           string    `json:"channelType" xorm:"index VARCHAR(30) notnull" validate:"required"`
	TransactionType       string    `json:"transactionType" xorm:"index VARCHAR(30) notnull"`
	OrderId               int64     `json:"orderId" xorm:"index notnull default 0"`
	RefundId              int64     `json:"refundId" xorm:"index notnull default 0"`
	StoreId               int64     `json:"storeId" xorm:"index notnull default 0"`
	ErrorType             string    `json:"errorType" xorm:"index VARCHAR(30)"`
	Error                 string    `json:"error" xorm:"TEXT"`
	Details               string    `json:"details" xorm:"VARCHAR(100)"`
	IsSuccess             bool      `json:"isSuccess" xorm:"index notnull default false"`
	OrderEntity           string    `json:"orderEntitys" xorm:"TEXT"`
	TransactionCreateDate time.Time `json:"transactionCreateDate" xorm:"index "`
	CreatedAt             time.Time `json:"createdAt" xorm:"index created"`
	UpdatedAt             time.Time `json:"updatedAt" xorm:"updated"`
}

func SaveSaleRecordSuccess(ctx context.Context, saleRecord *AssortedSaleRecord, isSuccess bool, err string, record interface{}) error {
	strEncoding, _ := json.Marshal(record)
	saleRecordSuccess := &SaleRecordSuccess{
		TenantCode:            saleRecord.TenantCode,
		ChannelType:           saleRecord.TransactionChannelType,
		TransactionType:       saleRecord.TransactionType,
		OrderId:               saleRecord.OrderId,
		RefundId:              saleRecord.RefundId,
		StoreId:               saleRecord.StoreId,
		IsSuccess:             isSuccess,
		Error:                 err,
		OrderEntity:           string(strEncoding),
		TransactionCreateDate: saleRecord.TransactionCreateDate,
	}
	if !isSuccess {
		errType, errMsg, details := ErrorTypeMessage(err)
		saleRecordSuccess.ErrorType = errType
		saleRecordSuccess.Error = errMsg
		saleRecordSuccess.Details = details
	}
	if err := saleRecordSuccess.Save(ctx); err != nil {
		return err
	}
	if !isSuccess {
		go EventPublish{}.PublishSaleFailEvent(ctx, saleRecordSuccess)
	}
	return nil
}

func ErrorTypeMessage(err string) (string, string, string) {
	switch err {
	case string(ItemFeeRateError):
		return err, "Item FeeRate not avalable 0", "正常扣率为0！"
	case string(CreatedIdError):
		return err, "CreatedId not avalable 0", "登录人员信息错误！"
	case string(POSSalesmanIdError):
		return err, "POS then SalesmanId not avalable 0", "销售人员信息错误！"
	case string(StoreIdError):
		return err, "StoreId not avalable 0", "卖场代码为空！"
	case string(TotalPriceError):
		return err, "TotalPrice not equals sum dtl price", "总金额计算错误！"
	case string(DiscountPriceError):
		return err, "DiscountPrice not correct", "折扣金额计算错误！"
	case string(DiscountPriceNotMatchOfferError):
		return err, "DiscountPrice not equals sum cartoffer's price", "折扣金额和促销金额不匹配！"
	case string(OfferNoError):
		return err, "OfferNo not avalable", "促销信息异常！"
	case string(StoreNotExistError):
		return err, "Store is not exists", "卖场信息异常！"
	case string(BrandNotMatchError):
		return err, "Store and Brand can't match", "卖场品牌信息不匹配！"
	case string(SkuNotExistError):
		return err, "Sku is not exists", "商品信息异常！"
	case string(ProductNotMatchError):
		return err, "Sku and Product can't match", "商品信息不匹配！"
	case string(SkuListPriceError):
		return err, "Sku's listprice not correct", "商品吊牌金额不匹配！"
	case string(ProductBrandError):
		return err, "Product's brand is not correct", "商品品牌信息不匹配！"
	case string(DistributedCashPriceError):
		return err, "Distributed CashPrice is not correct", "商品实付金额计算错误！"
	case string(TotalDistributedPaymentPriceError):
		return err, "Distributed PaymentPrice is not correct", "商品支付金额计算错误！"
	case string(MileageError):
		return err, "Mileage is not correct", "积分计算错误！"
	case string(PayMentNoExistError):
		return err, "PayMent not exists", "支付信息不存在！"
	default:
		return string(SaleRecordError), err, "上传数据处理异常！"
	}
}

func SaveOrderEventError(ctx context.Context, err string, event Event) error {
	order := event.Payload
	strEncoding, _ := json.Marshal(event)
	errType, errMsg, details := ErrorTypeMessage(err)
	saleRecordSuccess := &SaleRecordSuccess{
		OrderId:               order.Id,
		RefundId:              0,
		StoreId:               order.StoreId,
		TenantCode:            order.TenantCode,
		ChannelType:           order.SaleType,
		TransactionType:       "PLUS",
		IsSuccess:             false,
		ErrorType:             errType,
		Error:                 errMsg,
		Details:               details,
		OrderEntity:           string(strEncoding),
		TransactionCreateDate: order.CreatedAt,
	}

	if err := saleRecordSuccess.Save(ctx); err != nil {
		return err
	}
	go EventPublish{}.PublishSaleFailEvent(ctx, saleRecordSuccess)

	return nil
}
func SaveRefundEventError(ctx context.Context, err string, event Event) error {
	refund := event.Payload.Refunds[0]
	strEncoding, _ := json.Marshal(event)
	errType, errMsg, details := ErrorTypeMessage(err)
	saleRecordSuccess := &SaleRecordSuccess{
		OrderId:               event.Payload.Id,
		RefundId:              refund.Id,
		StoreId:               refund.StoreId,
		TenantCode:            refund.TenantCode,
		ChannelType:           refund.RefundType,
		TransactionType:       "MINUS",
		IsSuccess:             false,
		ErrorType:             errType,
		Error:                 errMsg,
		Details:               details,
		OrderEntity:           string(strEncoding),
		TransactionCreateDate: refund.CreatedAt,
	}

	if err := saleRecordSuccess.Save(ctx); err != nil {
		return err
	}
	go EventPublish{}.PublishSaleFailEvent(ctx, saleRecordSuccess)

	return nil
}

func (record *SaleRecordSuccess) Save(ctx context.Context) error {
	has, err := factory.DB(ctx).Exist(&SaleRecordSuccess{
		OrderId:         record.OrderId,
		RefundId:        record.RefundId,
		TenantCode:      record.TenantCode,
		ChannelType:     record.ChannelType,
		TransactionType: record.TransactionType,
	})
	if err != nil {
		return err
	}
	if has {
		if err := record.Update(ctx); err != nil {
			return err
		}
	} else {
		if err := record.Insert(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (record *SaleRecordSuccess) Update(ctx context.Context) error {
	if _, err := factory.DB(ctx).Where("order_id = ?", record.OrderId).
		And("refund_id = ?", record.RefundId).And("channel_type = ?", record.ChannelType).
		And("transaction_type = ?", record.TransactionType).
		And("tenant_code = ?", record.TenantCode).Cols("is_success", "error", "error_type", "order_entity", "store_id").
		Update(record); err != nil {
		return err
	}
	return nil
}

func (record *SaleRecordSuccess) Insert(ctx context.Context) error {
	if _, err := factory.DB(ctx).Insert(record); err != nil {
		return err
	}
	return nil
}

func (SaleRecordSuccess) GetSaleRecordLogs(ctx context.Context, searchInput SaleRecordLogSearchInput) (*SearchSaleRecordLogOutPut, error) {
	queryBulder := func() xorm.Interface {
		q := factory.DB(ctx).Where("1=1").And("is_success = ?", searchInput.IsSuccess)
		if searchInput.StoreId > 0 {
			q.And("store_id = ?", searchInput.StoreId)
		}
		if searchInput.OrderId > 0 {
			return q.And("order_id = ?", searchInput.OrderId)
		}
		if searchInput.RefundId > 0 {
			return q.And("refund_id = ?", searchInput.RefundId)
		}
		if searchInput.ErrType != "" {
			return q.And("error_type = ?", searchInput.ErrType)
		}
		if searchInput.ChannelType != "" {
			return q.And("channel_type = ?", searchInput.ChannelType)
		}
		if searchInput.TransactionType != "" {
			return q.And("transaction_type = ?", searchInput.TransactionType)
		}
		return q.And("created_at>=?", searchInput.StartTime).And("created_at<?", searchInput.EndTime.AddDate(0, 0, 1))
	}

	query := queryBulder()
	if searchInput.MaxResultCount > 0 {
		query.Limit(searchInput.MaxResultCount, searchInput.SkipCount)
	}
	var saleRecordLogs []SaleRecordSuccess
	totalCount, err := query.Desc("id").FindAndCount(&saleRecordLogs)
	if err != nil {
		return nil, err
	}
	return &SearchSaleRecordLogOutPut{
		TotalCount: totalCount,
		Items:      saleRecordLogs,
	}, nil
}
