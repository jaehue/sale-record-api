package models

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/hublabs/common/auth"
	"github.com/hublabs/sale-record-api/factory"

	"github.com/go-xorm/xorm"
	"github.com/pangpanglabs/goutils/behaviorlog"
)

func GetSaleRecordByOrderId(ctx context.Context, orderId int64, channelType, transactionType string) (*AssortedSaleRecord, error) {
	userClaim := UserClaim(auth.UserClaim{}.FromCtx(ctx))
	saleRecord := AssortedSaleRecord{OrderId: orderId, TenantCode: userClaim.tenantCode(), TransactionChannelType: channelType, TransactionType: transactionType}
	if has, err := factory.DB(ctx).Get(&saleRecord); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	saleRecord.getSaleRecordItemAndOffers(ctx)
	return &saleRecord, nil
}

func GetSaleRecordByRefundId(ctx context.Context, refundId int64, channelType, transactionType string) (*AssortedSaleRecord, error) {
	userClaim := UserClaim(auth.UserClaim{}.FromCtx(ctx))
	saleRecord := AssortedSaleRecord{RefundId: refundId, TenantCode: userClaim.tenantCode(), TransactionChannelType: channelType, TransactionType: transactionType}
	if has, err := factory.DB(ctx).Get(&saleRecord); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	saleRecord.getSaleRecordItemAndOffers(ctx)
	return &saleRecord, nil
}

func GetSaleRecordByTransactionId(ctx context.Context, transactionId int64) (*AssortedSaleRecord, error) {
	saleRecord := AssortedSaleRecord{TransactionId: transactionId}
	if has, err := factory.DB(ctx).Get(&saleRecord); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	saleRecord.getSaleRecordItemAndOffers(ctx)
	return &saleRecord, nil
}

func GetSaleRecords(ctx context.Context, saleRecordSearchInput SaleRecordSearchInput) (*SearchSaleRecordOutput, error) {
	if saleRecordSearchInput.TransactionId > 0 {
		saleRecord, err := GetSaleRecordByTransactionId(ctx, saleRecordSearchInput.TransactionId)
		if err != nil {
			return nil, err
		}
		return &SearchSaleRecordOutput{
			TotalCount: 1,
			Items: []AssortedSaleRecord{
				*saleRecord,
			},
		}, nil
	}
	queryBuilder := func() xorm.Interface {
		q := factory.DB(ctx).Where("1=1")
		if !saleRecordSearchInput.OrderStartTime.IsZero() && !saleRecordSearchInput.OrderEndTime.IsZero() {
			q.And("created>=?", saleRecordSearchInput.OrderStartTime).
				And("created<?", saleRecordSearchInput.OrderEndTime.AddDate(0, 0, 1))
		}
		if saleRecordSearchInput.OrderId > 0 {
			q.And("order_id =?", saleRecordSearchInput.OrderId)
		}
		if saleRecordSearchInput.RefundId > 0 {
			q.And("refund_id =?", saleRecordSearchInput.RefundId)
		}
		if saleRecordSearchInput.OrderIds != "" {
			idsArry := strings.Split(saleRecordSearchInput.OrderIds, ",")
			orderIds := []int64{}
			for _, item := range idsArry {
				id, _ := strconv.ParseInt(item, 10, 64)
				orderIds = append(orderIds, id)
			}
			q.In("order_id", orderIds)
		}
		if saleRecordSearchInput.RefundIds != "" {
			idsArry := strings.Split(saleRecordSearchInput.RefundIds, ",")
			refundIds := []int64{}
			for _, item := range idsArry {
				id, _ := strconv.ParseInt(item, 10, 64)
				refundIds = append(refundIds, id)
			}
			q.In("refund_id", refundIds)
		}
		if saleRecordSearchInput.CustomerId > -1 {
			q.And("customer_id =?", saleRecordSearchInput.CustomerId)
		}

		if saleRecordSearchInput.CreatedId > -1 {
			q.And("created_by =?", saleRecordSearchInput.CreatedId)
		}

		if saleRecordSearchInput.TransactionStatus != "" {
			q.And("transaction_status =?", saleRecordSearchInput.TransactionStatus)
		}

		if saleRecordSearchInput.TransactionType != "" {
			q.And("transaction_type =?", saleRecordSearchInput.TransactionType)
		}

		if saleRecordSearchInput.ChannelType != "" {
			q.And("transaction_channel_type =?", saleRecordSearchInput.ChannelType)
		}

		if saleRecordSearchInput.SalesmanId > -1 {
			q.And("salesman_id =?", saleRecordSearchInput.SalesmanId)
		}

		if saleRecordSearchInput.EmpId != "" {
			q.And("emp_id =?", saleRecordSearchInput.EmpId)
		}

		if saleRecordSearchInput.StoreId > -1 {
			q.And("store_id =?", saleRecordSearchInput.StoreId)
		}

		if saleRecordSearchInput.OuterOrderNo != "" {
			q.And("outer_order_no =?", saleRecordSearchInput.OuterOrderNo)
		}
		if saleRecordSearchInput.IsOutPaid != "" {
			isOutPaid, _ := strconv.ParseBool(saleRecordSearchInput.IsOutPaid)
			q.And("is_out_paid =?", isOutPaid)
		}
		return q
	}

	query := queryBuilder()

	if saleRecordSearchInput.MaxResultCount > 0 {
		query.Limit(saleRecordSearchInput.MaxResultCount, saleRecordSearchInput.SkipCount)
	}
	var saleRecords []AssortedSaleRecord
	totalCount, err := query.Desc("transaction_create_date").FindAndCount(&saleRecords)
	if err != nil {
		return nil, err
	}

	if saleRecords == nil {
		return nil, nil
	}

	for i, _ := range saleRecords {
		saleRecords[i].getSaleRecordItemAndOffers(ctx)
	}
	return &SearchSaleRecordOutput{
		TotalCount: totalCount,
		Items:      saleRecords,
	}, nil
}

func (saleRecord *AssortedSaleRecord) CreateSaleRecord(ctx context.Context) error {
	if row, err := factory.DB(ctx).Insert(saleRecord); err != nil {
		return err
	} else if row <= 0 {
		return errors.New("Insert SaleRecord not found.")
	}
	if len(saleRecord.SaleRecordDtls) > 0 {
		for i, _ := range saleRecord.SaleRecordDtls {
			saleRecordDtl := &saleRecord.SaleRecordDtls[i]
			saleRecordDtl.TransactionId = saleRecord.TransactionId
			if row, err := factory.DB(ctx).Insert(saleRecordDtl); err != nil {
				return err
			} else if row <= 0 {
				return errors.New("Insert SaleRecorddtl not found.")
			}
			for i, _ := range saleRecordDtl.AppliedSaleRecordCartOffers {
				saleRecordDtl.AppliedSaleRecordCartOffers[i].TransactionId = saleRecord.TransactionId
				saleRecordDtl.AppliedSaleRecordCartOffers[i].TransactionDtlId = saleRecordDtl.Id
				if row, err := factory.DB(ctx).Insert(saleRecordDtl.AppliedSaleRecordCartOffers[i]); err != nil {
					return err
				} else if row <= 0 {
					return errors.New("Insert SaleRecordItemAppliedCartOffer not found.")
				}
			}
		}
	}

	if len(saleRecord.AppliedSaleRecordCartOffers) > 0 {
		for i, _ := range saleRecord.AppliedSaleRecordCartOffers {
			saleRecord.AppliedSaleRecordCartOffers[i].TransactionId = saleRecord.TransactionId
		}
		if row, err := factory.DB(ctx).Insert(saleRecord.AppliedSaleRecordCartOffers); err != nil {
			return err
		} else if row <= 0 {
			return errors.New("Insert AppliedSaleRecordCartOffers not found.")
		}
	}

	var itemOffers = make([]AppliedSaleRecordItemOffer, 0)
	for _, saleRecordItem := range saleRecord.SaleRecordDtls {
		for _, itemOffer := range saleRecordItem.AppliedSaleRecordItemOffers {
			itemOffer.TransactionId = saleRecord.TransactionId
			itemOffers = append(itemOffers, itemOffer)
		}
	}

	if len(itemOffers) > 0 {
		if row, err := factory.DB(ctx).Insert(itemOffers); err != nil {
			return err
		} else if row <= 0 {
			return nil
		}
	}

	if len(saleRecord.Payments) > 0 {
		for i, _ := range saleRecord.Payments {
			saleRecord.Payments[i].TransactionId = saleRecord.TransactionId
		}

		if row, err := factory.DB(ctx).Insert(saleRecord.Payments); err != nil {
			return err
		} else if row <= 0 {
			return nil
		}
	}

	return nil
}

func (saleRecord *AssortedSaleRecord) getSaleRecordItemAndOffers(ctx context.Context) {
	saleRecord.SaleRecordDtls, _ = getSaleRecordItemByTransactionId(ctx, saleRecord.TransactionId)
	saleRecord.AppliedSaleRecordCartOffers, _ = getSaleRecordOffersByTransactionId(ctx, saleRecord.TransactionId)
	for i, t := range saleRecord.SaleRecordDtls {
		saleRecord.SaleRecordDtls[i].AppliedSaleRecordCartOffers, _ = getSaleRecordItemCartOffersBy(ctx, saleRecord.TransactionId, t.Id)
		if t.TotalDiscountPrice == 0 {
			continue
		}
		saleRecord.SaleRecordDtls[i].AppliedSaleRecordItemOffers, _ = getSaleRecordItemOffersBy(ctx, saleRecord.TransactionId, t.OrderItemId)
	}
	saleRecord.Payments, _ = getSaleRecordPaymentsByTransactionId(ctx, saleRecord.TransactionId)
}
func getSaleRecordItemByTransactionId(ctx context.Context, transactionId int64) ([]AssortedSaleRecordDtl, error) {
	var saleRecordItems []AssortedSaleRecordDtl
	if err := factory.DB(ctx).Where("transaction_id = ?", transactionId).
		Find(&saleRecordItems); err != nil {
		return nil, err
	}
	return saleRecordItems, nil
}

func getSaleRecordOffersByTransactionId(ctx context.Context, transactionId int64) ([]AppliedSaleRecordCartOffer, error) {
	var saleRecordOffers []AppliedSaleRecordCartOffer
	if err := factory.DB(ctx).Where("transaction_id = ?", transactionId).
		Find(&saleRecordOffers); err != nil {
		return nil, err
	}
	return saleRecordOffers, nil
}

func getSaleRecordPaymentsByTransactionId(ctx context.Context, transactionId int64) ([]AssortedSaleRecordPayment, error) {
	var postPayments []AssortedSaleRecordPayment
	if err := factory.DB(ctx).Where("transaction_id = ?", transactionId).
		Find(&postPayments); err != nil {
		return nil, err
	}
	return postPayments, nil
}

func getSaleRecordItemOffersBy(ctx context.Context, transactionId, transactionDtlId int64) ([]AppliedSaleRecordItemOffer, error) {
	var saleRecordItemOffers []AppliedSaleRecordItemOffer
	if err := factory.DB(ctx).Where("transaction_id = ? and transaction_dtl_id = ?", transactionId, transactionDtlId).
		Find(&saleRecordItemOffers); err != nil {
		return nil, err
	}
	return saleRecordItemOffers, nil
}
func getSaleRecordItemCartOffersBy(ctx context.Context, transactionId, transactionDtlId int64) ([]SaleRecordItemAppliedCartOffer, error) {
	var saleRecordItemCartOffers []SaleRecordItemAppliedCartOffer
	if err := factory.DB(ctx).Where("transaction_id = ? and transaction_dtl_id = ?", transactionId, transactionDtlId).
		Find(&saleRecordItemCartOffers); err != nil {
		return nil, err
	}
	return saleRecordItemCartOffers, nil
}

type AssortedSaleRecord struct {
	TransactionId int64  `json:"transactionId" xorm:"pk notnull autoincr"`
	EmpId         string `json:"empId" query:"empId" xorm:"index VARCHAR(50) notnull" validate:"required"`
	// EMALL offline shop salesman empid
	SalesmanEmpId       string  `json:"salesmanEmpId" query:"salesmanEmpId" xorm:"index VARCHAR(50)"`
	ChannelId           int64   `json:"channelId" query:"channelId" xorm:"index notnull" validate:"gte=0"`
	CustomerId          int64   `json:"customerId" query:"customerId" xorm:"index notnull" validate:"required"`
	DiscountCouponPrice float64 `json:"discountCouponPrice" query:"discountCouponPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	DiscountOfferPrice  float64 `json:"discountOfferPrice" query:"discountOfferPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	FreightPrice        float64 `json:"freightPrice" query:"freightPrice" xorm:"DECIMAL(18,2) default 0"`
	IsOutPaid           bool    `json:"isOutPaid" query:"isOutPaid" xorm:"index default false" `
	IsRefund            bool    `json:"isRefund" query:"isRefund" xorm:"index default false" `
	Mileage             float64 `json:"mileage" query:"mileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	MileagePrice        float64 `json:"mileagePrice" query:"mileagePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	CashPrice           float64 `json:"cashPrice" query:"cashPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	ObtainMileage       float64 `json:"obtainMileage" query:"obtainMileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	OrderId             int64   `json:"orderId" query:"orderId" xorm:"index"`
	OuterOrderNo        string  `json:"outerOrderNo" query:"outerOrderNo" xorm:"index"`
	RefundId            int64   `json:"refundId" xorm:"index"`
	SalesmanId          int64   `json:"salesmanId" query:"salesmanId" xorm:"index notnull" validate:"gte=0"`
	StoreId             int64   `json:"storeId" query:"storeId" xorm:"index notnull" validate:"gte=0"`
	// EMALL offline shop code
	SalesmanShopCode            string                       `json:"salesmanShopCode" query:"salesmanShopCode" xorm:"index VARCHAR(50)"`
	TenantCode                  string                       `json:"tenantCode" query:"tenantCode" xorm:"index VARCHAR(50) notnull" validate:"required"`
	TotalDiscountPrice          float64                      `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalListPrice              float64                      `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalSalePrice              float64                      `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalTransactionPrice       float64                      `json:"totalTransactionPrice" query:"totalTransactionPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TransactionChannelType      string                       `json:"transactionChannelType" query:"transactionChannelType" xorm:"index VARCHAR(30) notnull" validate:"required"`
	TransactionCreateDate       time.Time                    `json:"transactionCreateDate"`
	TransactionStatus           string                       `json:"transactionStatus" query:"transactionStatus" xorm:"index VARCHAR(30) notnull" validate:"required"`
	TransactionType             string                       `json:"transactionType" query:"transactionType" xorm:"index VARCHAR(30) notnull" validate:"required"`
	TransactionUpdateDate       time.Time                    `json:"transactionUpdateDate"`
	TransactionCreatedId        int64                        `json:"transactionCreatedId" query:"transactionCreatedId" xorm:""`
	BaseTrimCode                string                       `json:"baseTrimCode"`
	Created                     time.Time                    `json:"created"`
	CreatedBy                   string                       `json:"createdBy"`
	Modified                    time.Time                    `json:"modified"`
	ModifiedBy                  string                       `json:"modifiedBy"`
	SaleRecordDtls              []AssortedSaleRecordDtl      `json:"saleRecordDtls" xorm:"-"`
	AppliedSaleRecordCartOffers []AppliedSaleRecordCartOffer `json:"appliedSaleRecordCartOffers" xorm:"-"`
	Payments                    []AssortedSaleRecordPayment  `json:"payments" xorm:"-"`
}

type AssortedSaleRecordDtl struct {
	Id                             int64                            `json:"id" xorm:"pk notnull autoincr"`
	TransactionId                  int64                            `json:"transactionId" query:"transactionId" xorm:"index notnull" validate:"required"`
	OrderItemId                    int64                            `json:"orderItemId" query:"orderItemId" xorm:"index notnull" validate:"required"`
	RefundItemId                   int64                            `json:"refundItemId" query:"refundItemId" xorm:"index notnull default 0"`
	BrandId                        int64                            `json:"brandId" query:"brandId" xorm:"index notnull" validate:"gte=0"`
	BrandCode                      string                           `json:"brandCode" query:"brandCode" xorm:"index notnull" validate:"gte=0"`
	ItemCode                       string                           `json:"itemCode" query:"itemCode" xorm:"index notnull" validate:"required"`
	ItemName                       string                           `json:"itemName" query:"itemName" xorm:"notnull" validate:"required"`
	ProductId                      int64                            `json:"productId" query:"productId" xorm:"index notnull" validate:"gte=0"`
	SkuId                          int64                            `json:"skuId" query:"skuId" xorm:"notnull" validate:"gte=0"`
	SkuImg                         string                           `json:"skuImg" query:"skuImg" xorm:"VARCHAR(200)" validate:""`
	ItemFee                        float64                          `json:"itemFee" query:"itemFee" xorm:"DECIMAL(18,2) default 0.00" validate:"gte=0"`
	FeeRate                        float64                          `json:"FeeRate" query:"FeeRate" xorm:"DECIMAL(18,4) default 0.00" validate:"gte=0"`
	ListPrice                      float64                          `json:"listPrice" query:"listPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0" `
	SalePrice                      float64                          `json:"salePrice" query:"salePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Quantity                       int                              `json:"quantity" query:"quantity" xorm:"notnull default 0"`
	DistributedCashPrice           float64                          `json:"distributedCashPrice" query:"distributedCashPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDistributedCartOfferPrice float64                          `json:"totalDistributedCartOfferPrice" query:"totalDistributedCartOfferPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDistributedItemOfferPrice float64                          `json:"totalDistributedItemOfferPrice" query:"totalDistributedItemOfferPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDistributedPaymentPrice   float64                          `json:"totalDistributedPaymentPrice" query:"totalDistributedPaymentPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalListPrice                 float64                          `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalSalePrice                 float64                          `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDiscountPrice             float64                          `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalTransactionPrice          float64                          `json:"totalTransactionPrice" query:"totalTransactionPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Mileage                        float64                          `json:"mileage" query:"mileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	MileagePrice                   float64                          `json:"mileagePrice" query:"mileagePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	ObtainMileage                  float64                          `json:"obtainMileage" query:"obtainMileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	IsDelivery                     bool                             `json:"isDelivery" query:"isDelivery" xorm:"index default false"`
	Status                         string                           `json:"status" query:"status" xorm:"index VARCHAR(30) notnull" validate:"required"`
	Created                        time.Time                        `json:"created"`
	CreatedBy                      string                           `json:"createdBy"`
	Modified                       time.Time                        `json:"modified"`
	ModifiedBy                     string                           `json:"modifiedBy"`
	AppliedSaleRecordItemOffers    []AppliedSaleRecordItemOffer     `json:"appliedSaleRecordItemOffers" xorm:"-"`
	AppliedSaleRecordCartOffers    []SaleRecordItemAppliedCartOffer `json:"appliedSaleRecordCartOffers" xorm:"-"`
}

type AppliedSaleRecordCartOffer struct {
	Id            int64   `json:"id" xorm:"pk notnull autoincr"`
	TenantCode    string  `json:"tenantCode"`
	OfferNo       string  `json:"offerNo"`
	CouponNo      string  `json:"couponNo"`
	ItemIds       string  `json:"itemIds" xorm:"TEXT"`
	TargetItemIds string  `json:"targetItemIds" xorm:"TEXT"`
	Price         float64 `json:"price"`
	Type          string  `json:"type"`
	TransactionId int64   `json:"transactionId" query:"transactionId" xorm:"index notnull" validate:"required"`
	TargetType    string  `josn:"targetType" xorm:"VARCHAR(10)"`
}

type SaleRecordItemAppliedCartOffer struct {
	Id               int64   `json:"id" xorm:"pk notnull autoincr"`
	OfferNo          string  `json:"offerNo"`
	CouponNo         string  `json:"couponNo"`
	TransactionId    int64   `json:"transactionId" query:"transactionId" xorm:"index notnull" validate:"required"`
	TransactionDtlId int64   `json:"transactionDtlId" xorm:"index notnull"  validate:"required"`
	Type             string  `json:"type"`
	TargetType       string  `josn:"targetType"`
	IsTarget         bool    `josn:"isTarget"`
	Price            float64 `json:"price"`
}
type AppliedSaleRecordItemOffer struct {
	Id               int64   `json:"id" xorm:"pk notnull autoincr"`
	TenantCode       string  `json:"tenantCode"`
	OfferNo          string  `json:"offerNo"`
	CouponNo         string  `json:"couponNo"`
	ItemCodes        string  `json:"itemCodes"`
	Price            float64 `json:"price"`
	Type             string  `json:"type"`
	TransactionId    int64   `json:"transactionId" query:"transactionId" xorm:"index notnull" validate:"required"`
	TargetType       string  `josn:"targetType"`
	ItemCode         string  `json:"itemCode"`
	TransactionDtlId int64   `json:"transactionDtlId"`
}
type AssortedSaleRecordPayment struct {
	Id            int64     `json:"id" xorm:"pk notnull autoincr"`
	TransactionId int64     `json:"transactionId" query:"transactionId" xorm:"index notnull" validate:"required"`
	SeqNo         int64     `json:"seqNo"`
	PayMethod     string    `json:"payMethod"`
	PayAmt        float64   `json:"payAmt"`
	CreatedAt     time.Time `json:"CreatedBy"`
}

func (saleRecord *AssortedSaleRecord) translateEventMessage(ctx context.Context) map[string]interface{} {
	saleRecordEvent := saleRecord.MakeSaleRecordEvent()

	return map[string]interface{}{
		"requestId": behaviorlog.FromCtx(ctx).RequestID,
		"actionId":  behaviorlog.FromCtx(ctx).ActionID,
		"authToken": behaviorlog.FromCtx(ctx).AuthToken,
		"payload":   saleRecordEvent,
	}
}

func (saleRecord *AssortedSaleRecord) MakeSaleRecordEvent() SaleRecordEvent {
	var saleRecordEvent SaleRecordEvent
	saleRecordEvent.TransactionId = saleRecord.TransactionId
	saleRecordEvent.AssortedSaleRecordDtlList = saleRecord.MakeSaleRecordDtlEvent()
	saleRecordEvent.TenantCode = saleRecord.TenantCode
	saleRecordEvent.StoreId = saleRecord.StoreId
	saleRecordEvent.OrderId = saleRecord.OrderId
	saleRecordEvent.OuterOrderNo = saleRecord.OuterOrderNo
	saleRecordEvent.RefundId = saleRecord.RefundId
	saleRecordEvent.TransactionType = saleRecord.TransactionType
	saleRecordEvent.TransactionChannelType = saleRecord.TransactionChannelType
	saleRecordEvent.TransactionStatus = saleRecord.TransactionStatus
	saleRecordEvent.TransactionCreateDate = saleRecord.TransactionCreateDate
	saleRecordEvent.TransactionUpdateDate = saleRecord.TransactionUpdateDate
	saleRecordEvent.CustomerId = saleRecord.CustomerId
	saleRecordEvent.EmpId = saleRecord.EmpId
	saleRecordEvent.SalesmanId = saleRecord.SalesmanId
	saleRecordEvent.SalesmanEmpId = saleRecord.SalesmanEmpId
	saleRecordEvent.SalesmanShopCode = saleRecord.SalesmanShopCode
	saleRecordEvent.TotalPrice = TotalPriceEvent{
		ListPrice:        saleRecord.TotalListPrice,
		SalePrice:        saleRecord.TotalSalePrice,
		DiscountPrice:    saleRecord.TotalDiscountPrice,
		TransactionPrice: saleRecord.TotalTransactionPrice,
	}
	saleRecordEvent.FreightPrice = saleRecord.FreightPrice
	saleRecordEvent.Mileage = saleRecord.Mileage
	saleRecordEvent.MileagePrice = saleRecord.MileagePrice
	saleRecordEvent.ObtainMileage = saleRecord.ObtainMileage
	saleRecordEvent.CashPrice = saleRecord.CashPrice
	saleRecordEvent.IsOutPaid = saleRecord.IsOutPaid
	saleRecordEvent.CartOffers = saleRecord.MakeCartOffers()
	saleRecordEvent.Committed = CommittedEvent{
		Created:   saleRecord.Created,
		CreatedBy: saleRecord.CreatedBy,
	}
	saleRecordEvent.BaseTrimCode = saleRecord.BaseTrimCode
	saleRecordEvent.TransactionCreatedId = saleRecord.TransactionCreatedId
	saleRecordEvent.Payments = saleRecord.MakePostPaymentEvent()
	return saleRecordEvent
}

func (saleRecord *AssortedSaleRecord) MakeSaleRecordDtlEvent() []SaleRecordDtlEvent {
	var saleRecordDtlEvents = make([]SaleRecordDtlEvent, 0)
	for _, dtl := range saleRecord.SaleRecordDtls {
		saleRecodeDtlEvent := SaleRecordDtlEvent{
			Id:            dtl.Id,
			OrderItemId:   dtl.OrderItemId,
			RefundItemId:  dtl.RefundItemId,
			BrandId:       dtl.BrandId,
			BrandCode:     dtl.BrandCode,
			ItemCode:      dtl.ItemCode,
			ItemName:      dtl.ItemName,
			ProductId:     dtl.ProductId,
			SkuId:         dtl.SkuId,
			SkuImg:        dtl.SkuImg,
			ListPrice:     dtl.ListPrice,
			SalePrice:     dtl.SalePrice,
			Quantity:      dtl.Quantity,
			Mileage:       dtl.Mileage,
			MileagePrice:  dtl.MileagePrice,
			ObtainMileage: dtl.ObtainMileage,
			TotalPrice: TotalPriceEvent{
				ListPrice:        dtl.TotalListPrice,
				SalePrice:        dtl.TotalSalePrice,
				DiscountPrice:    dtl.TotalDiscountPrice,
				TransactionPrice: dtl.TotalTransactionPrice,
			},
			DistributedPrice: DistributedPriceEvent{
				TotalDistributedItemOfferPrice: dtl.TotalDistributedItemOfferPrice,
				TotalDistributedCartOfferPrice: dtl.TotalDistributedCartOfferPrice,
				TotalDistributedPaymentPrice:   dtl.TotalDistributedPaymentPrice,
				DistributedCashPrice:           dtl.DistributedCashPrice,
			},
			Status:     dtl.Status,
			FeeRate:    dtl.FeeRate,
			ItemOffers: dtl.MakeItemOffers(),
			CartOffers: dtl.MakeCartOffers(),
			Committed: CommittedEvent{
				Created:   dtl.Created,
				CreatedBy: dtl.CreatedBy,
			},
		}
		saleRecordDtlEvents = append(saleRecordDtlEvents, saleRecodeDtlEvent)
	}
	return saleRecordDtlEvents
}

func (saleRecord *AssortedSaleRecord) MakeCartOffers() []CartOfferEvent {
	var cartOffers = make([]CartOfferEvent, 0)
	for _, offer := range saleRecord.AppliedSaleRecordCartOffers {
		cartOffer := CartOfferEvent{
			OfferId:       offer.Id,
			OfferNo:       offer.OfferNo,
			CouponNo:      offer.CouponNo,
			TargetItemIds: offer.TargetItemIds,
			ItemIds:       offer.ItemIds,
			Price:         offer.Price,
		}
		cartOffers = append(cartOffers, cartOffer)
	}
	return cartOffers
}

func (saleRecordDtl AssortedSaleRecordDtl) MakeItemOffers() []ItemOfferEvent {
	var itemOffers = make([]ItemOfferEvent, 0)
	for _, offer := range saleRecordDtl.AppliedSaleRecordItemOffers {
		itemOffer := ItemOfferEvent{
			OfferId:   offer.Id,
			OfferNo:   offer.OfferNo,
			CouponNo:  offer.CouponNo,
			ItemCodes: offer.ItemCodes,
			ItemCode:  offer.ItemCode,
			Price:     offer.Price,
		}
		itemOffers = append(itemOffers, itemOffer)
	}
	return itemOffers
}

func (saleRecordDtl AssortedSaleRecordDtl) MakeCartOffers() []ItemCartOfferEvent {
	var cartOffers = make([]ItemCartOfferEvent, 0)
	for _, offer := range saleRecordDtl.AppliedSaleRecordCartOffers {
		cartOffer := ItemCartOfferEvent{
			OfferNo:    offer.OfferNo,
			CouponNo:   offer.CouponNo,
			TargetType: offer.TargetType,
			IsTarget:   offer.IsTarget,
			Type:       offer.Type,
			Price:      offer.Price,
		}
		cartOffers = append(cartOffers, cartOffer)
	}
	return cartOffers
}

func (saleRecord *AssortedSaleRecord) MakePostPaymentEvent() []PaymentEvent {
	var paymentEvents = make([]PaymentEvent, 0)
	for _, payment := range saleRecord.Payments {
		paymentEvent := PaymentEvent{
			SeqNo:     payment.SeqNo,
			PayMethod: payment.PayMethod,
			PayAmt:    payment.PayAmt,
			CreatedAt: payment.CreatedAt,
		}
		paymentEvents = append(paymentEvents, paymentEvent)
	}
	return paymentEvents
}
