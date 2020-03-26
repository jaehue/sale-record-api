package models

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/pangpanglabs/goutils/number"
)

type OrderEventHandler struct{}

func (o OrderEventHandler) Handle(ctx context.Context, event Event) error {
	if orderType := FindOrderType(event.Status); orderType.CanMakeSaleRecord == false {
		return nil
	}
	var (
		saleRecord *AssortedSaleRecord
		err        error
		exists     bool
	)

	if event.Payload.Refunds != nil {
		if exists, saleRecord, err = (RefundEvent{}).HandleRefundEvent(ctx, event); err != nil {
			if saveErr := SaveRefundEventError(ctx, err.Error(), event); saveErr != nil {
				return saveErr
			}
			return nil
		}
	} else {
		if exists, saleRecord, err = handleOrderEvent(ctx, event); err != nil {
			if saveErr := SaveOrderEventError(ctx, err.Error(), event); saveErr != nil {
				return saveErr
			}
			return nil
		}
	}
	if exists {
		go EventPublish{}.PublishSaleRecordEvent(ctx, saleRecord)
		return nil
	}

	if err := saleRecord.CreateSaleRecord(ctx); err != nil {
		if saveErr := SaveSaleRecordSuccess(ctx, saleRecord, false, err.Error(), event); saveErr != nil {
			return saveErr
		}
		return nil
	}
	go EventPublish{}.PublishSaleRecordEvent(ctx, saleRecord)

	if saveErr := SaveSaleRecordSuccess(ctx, saleRecord, true, "", saleRecord); saveErr != nil {
		return saveErr
	}

	return nil
}

func CheckSaleRecordExits(ctx context.Context, orderId, refundId int64, channelType, transactionType string) (bool, *AssortedSaleRecord, error) {
	if refundId > 0 {
		if transactionType == "" {
			transactionType = "MINUS"
		}
		if saleRecord, err := GetSaleRecordByRefundId(ctx, refundId, channelType, transactionType); err != nil {
			return false, nil, err
		} else if saleRecord != nil {
			return true, saleRecord, nil
		}
	} else {
		if transactionType == "" {
			transactionType = "PLUS"
		}
		if saleRecord, err := GetSaleRecordByOrderId(ctx, orderId, channelType, transactionType); err != nil {
			return false, nil, err
		} else if saleRecord != nil {
			return true, saleRecord, nil
		}
	}
	return false, nil, nil
}

func handleOrderEvent(ctx context.Context, event Event) (bool, *AssortedSaleRecord, error) {
	var payload = event.Payload
	if exists, saleRecord, err := CheckSaleRecordExits(ctx, payload.Id, 0, payload.SaleType, ""); err != nil {
		return false, nil, err
	} else if exists {
		saleRecord.TransactionStatus = event.Status
		return true, saleRecord, nil
	}

	if err := event.Payload.OrderValidate(ctx); err != nil {
		return false, nil, err
	}

	numberSetting, baseTrimCode, err := GetRoundSetting(ctx, event.Payload.StoreId)
	if err != nil {
		return false, nil, err
	}

	saleRecord, err := event.Payload.makeAssortedSaleRecord(ctx, numberSetting)
	if err != nil {
		return false, nil, err
	}
	saleRecord.BaseTrimCode = baseTrimCode
	return false, saleRecord, nil
}

type Event struct {
	EntityType string     `json:"entityType"` // Order, Refund, OrderDelivery, RefundDelivery, StockDistribution
	Status     string     `json:"status"`
	Payload    OrderEvent `json:"payload"`
}

type OrderEvent struct {
	Id                 int64            `json:"id,omitempty"`
	OuterOrderNo       string           `json:"outerOrderNo"`
	TenantCode         string           `json:"tenantCode"`
	StoreId            int64            `json:"storeId"`
	ChannelId          int64            `json:"channelId" `
	SaleType           string           `json:"saleType" `
	CustomerId         int64            `json:"customerId"`
	SalesmanId         int64            `json:"salesmanId"`
	TotalListPrice     float64          `json:"totalListPrice"`
	TotalSalePrice     float64          `json:"totalSalePrice"`
	TotalDiscountPrice float64          `json:"totalDiscountPrice"`
	FreightPrice       float64          `json:"freightPrice"`
	TotalPaymentPrice  float64          `json:"totalPaymentPrice"`
	Mileage            float64          `json:"mileage"`
	MileagePrice       float64          `json:"mileagePrice"`
	ObtainMileage      float64          `json:"obtainMileage"`
	CashPrice          float64          `json:"cashPrice"`
	IsOutPaid          bool             `json:"isOutPaid"`
	Status             string           `json:"status"`
	Items              []OrderEventItem `json:"items,omitempty"`
	Offers             []OfferEvent     `json:"offers,omitempty"`
	Refunds            []RefundEvent    `json:"refunds,omitempty"`
	CreatedAt          time.Time        `json:"createdAt,omitempty" xorm:"created"`
	UpdatedAt          time.Time        `json:"updatedAt,omitempty" xorm:"updated"`
	CreatedId          int64            `json:"createdId"`
}

type OfferEvent struct {
	OfferNo       string  `json:"offerNo"`
	CouponNo      string  `json:"couponNo"`
	ItemIds       string  `json:"itemIds"`
	TargetItemIds string  `json:"targetItemIds"`
	Price         float64 `json:"price"`
	TargetType    string  `json:"targetType"`
}
type OrderEventItem struct {
	Id                             int64                 `json:"id,omitempty"`
	OuterOrderItemNo               string                `json:"outerOrderItemNo"`
	ItemCode                       string                `json:"itemCode"`
	ItemName                       string                `json:"itemName"`
	ItemFee                        float64               `json:"itemFee"`
	FeeRate                        float64               `json:"feeRate"`
	ProductId                      int64                 `json:"productId"`
	SkuId                          int64                 `json:"skuId"`
	SkuImg                         string                `json:"skuImg"`
	Option                         string                `json:"option"`
	ListPrice                      float64               `json:"listPrice"`
	SalePrice                      float64               `json:"salePrice"`
	Quantity                       int                   `json:"quantity"`
	TotalListPrice                 float64               `json:"totalListPrice"`
	TotalSalePrice                 float64               `json:"totalSalePrice"`
	TotalDiscountPrice             float64               `json:"totalDiscountPrice"`
	TotalPaymentPrice              float64               `json:"totalPaymentPrice"`
	Mileage                        float64               `json:"mileage"`
	MileagePrice                   float64               `json:"mileagePrice"`
	ObtainMileage                  float64               `json:"obtainMileage"`
	CashPrice                      float64               `json:"cashPrice"`
	Status                         string                `json:"status"`
	IsDelivery                     bool                  `json:"isDelivery" `
	IsStockChecked                 bool                  `json:"isStockChecked" `
	GroupOffers                    []OrderItemOfferEvent `json:"groupOffers,omitempty"`
	CreatedAt                      time.Time             `json:"createdAt"`
	UpdatedAt                      time.Time             `json:"updatedAt"`
	TotalDistributedCartOfferPrice float64               `json:"totalDistributedCartOfferPrice"`

	// TotalDistributedCartOfferPrice float64
}
type OrderItemOfferEvent struct {
	OfferNo    string  `json:"offerNo"`
	CouponNo   string  `json:"couponNo"`
	TargetType string  `json:"targetType"`
	IsTarget   bool    `json:"isTarget" `
	Price      float64 `json:"price"`
}

func (orderEvent *OrderEvent) makeAssortedSaleRecord(ctx context.Context, setting *number.Setting) (*AssortedSaleRecord, error) {
	// orderEvent.distributeCartOfferPrice(setting)
	var saleRecode = AssortedSaleRecord{}
	saleRecode.OrderId = orderEvent.Id
	saleRecode.TenantCode = orderEvent.TenantCode
	saleRecode.CashPrice = orderEvent.CashPrice
	saleRecode.ChannelId = orderEvent.ChannelId
	saleRecode.Created = time.Now().UTC()
	saleRecode.CreatedBy = "kafka-listener"
	saleRecode.Modified = time.Now().UTC()
	saleRecode.ModifiedBy = "kafka-listener"
	saleRecode.CustomerId = orderEvent.CustomerId
	saleRecode.RefundId = 0
	saleRecode.IsRefund = false
	saleRecode.DiscountOfferPrice, saleRecode.DiscountCouponPrice = orderEvent.getDiscountPrice(setting)
	saleRecode.TotalTransactionPrice = orderEvent.TotalPaymentPrice
	saleRecode.TransactionChannelType = orderEvent.SaleType
	saleRecode.TransactionCreateDate = orderEvent.CreatedAt
	saleRecode.TransactionStatus = orderEvent.Status
	saleRecode.TransactionType = "PLUS"
	saleRecode.TransactionUpdateDate = orderEvent.CreatedAt
	saleRecode.FreightPrice = orderEvent.FreightPrice
	saleRecode.IsOutPaid = orderEvent.IsOutPaid
	saleRecode.Mileage = orderEvent.Mileage
	saleRecode.MileagePrice = orderEvent.MileagePrice
	saleRecode.ObtainMileage = orderEvent.ObtainMileage
	saleRecode.OuterOrderNo = orderEvent.OuterOrderNo
	saleRecode.SalesmanId = orderEvent.SalesmanId
	saleRecode.StoreId = orderEvent.StoreId
	saleRecode.TenantCode = orderEvent.TenantCode
	saleRecode.TotalDiscountPrice = orderEvent.TotalDiscountPrice
	saleRecode.TotalListPrice = orderEvent.TotalListPrice
	saleRecode.TotalSalePrice = orderEvent.TotalSalePrice
	saleRecode.TransactionCreatedId = orderEvent.CreatedId
	saleRecode.AppliedSaleRecordCartOffers = orderEvent.makeSaleRecordCartOffers()

	var err error
	if saleRecode.SaleRecordDtls, err = orderEvent.makeAssortedSaleRecordDtls(ctx, setting); err != nil {
		return nil, err
	}
	if saleRecode.Payments, err = orderEvent.makePayments(ctx); err != nil {
		return nil, err
	}
	empId, err := orderEvent.makeEmpId(ctx)
	if err != nil {
		return nil, err
	}
	saleRecode.EmpId = empId
	return &saleRecode, nil
}

func (orderEvent *OrderEvent) makeAssortedSaleRecordDtls(ctx context.Context, setting *number.Setting) ([]AssortedSaleRecordDtl, error) {
	var saleRecodeDtls = make([]AssortedSaleRecordDtl, 0)
	for _, orderItemEvent := range orderEvent.Items {
		item, err := Item{}.GetItemByCode(ctx, orderItemEvent.ItemCode)
		if err != nil {
			return nil, err
		}
		saleRecordDtl := AssortedSaleRecordDtl{
			BrandCode:                      item.Sku.Product.Brand.Code,
			BrandId:                        item.Sku.Product.Brand.Id,
			Created:                        orderItemEvent.CreatedAt,
			CreatedBy:                      "kafka-listener",
			Modified:                       orderItemEvent.UpdatedAt,
			ModifiedBy:                     "kafka-listener",
			DistributedCashPrice:           number.ToFixed(orderItemEvent.TotalPaymentPrice-orderItemEvent.TotalDistributedCartOfferPrice-orderItemEvent.MileagePrice, nil),
			TotalDistributedCartOfferPrice: orderItemEvent.TotalDistributedCartOfferPrice,
			TotalDistributedItemOfferPrice: number.ToFixed(orderItemEvent.TotalListPrice-orderItemEvent.TotalSalePrice, nil),
			TotalDistributedPaymentPrice:   number.ToFixed(orderItemEvent.TotalPaymentPrice-orderItemEvent.TotalDistributedCartOfferPrice, nil),
			IsDelivery:                     orderItemEvent.IsDelivery,
			ItemCode:                       orderItemEvent.ItemCode,
			ItemName:                       orderItemEvent.ItemName,
			ItemFee:                        orderItemEvent.ItemFee,
			FeeRate:                        orderItemEvent.FeeRate,
			ListPrice:                      orderItemEvent.ListPrice,
			OrderItemId:                    orderItemEvent.Id,
			ProductId:                      orderItemEvent.ProductId,
			Quantity:                       orderItemEvent.Quantity,
			SalePrice:                      orderItemEvent.SalePrice,
			SkuId:                          orderItemEvent.SkuId,
			SkuImg:                         orderItemEvent.SkuImg,
			Status:                         orderItemEvent.Status,
			TotalDiscountPrice:             orderItemEvent.TotalDiscountPrice,
			TotalListPrice:                 orderItemEvent.TotalListPrice,
			TotalSalePrice:                 orderItemEvent.TotalSalePrice,
			TotalTransactionPrice:          orderItemEvent.TotalPaymentPrice,
			Mileage:                        orderItemEvent.Mileage,
			MileagePrice:                   orderItemEvent.MileagePrice,
			ObtainMileage:                  orderItemEvent.ObtainMileage,
			AppliedSaleRecordItemOffers:    orderItemEvent.makeSaleRecordItemOffers(item, orderEvent.TenantCode),
			AppliedSaleRecordCartOffers:    orderItemEvent.makeSaleRecordItemCartOffers(),
		}
		if saleRecordDtl.DistributedCashPrice < 0 {
			return nil, errors.New(string(DistributedCashPriceError))
		}
		if saleRecordDtl.TotalDistributedPaymentPrice < 0 {
			return nil, errors.New(string(TotalDistributedPaymentPriceError))
		}
		saleRecodeDtls = append(saleRecodeDtls, saleRecordDtl)
	}
	return saleRecodeDtls, nil
}

func (orderEvent *OrderEvent) makeSaleRecordCartOffers() []AppliedSaleRecordCartOffer {
	var appliedSaleRecordCartOffers = make([]AppliedSaleRecordCartOffer, 0)
	for _, offer := range orderEvent.Offers {
		var cartOffer = AppliedSaleRecordCartOffer{
			TenantCode:    orderEvent.TenantCode,
			OfferNo:       offer.OfferNo,
			CouponNo:      offer.CouponNo,
			ItemIds:       offer.ItemIds,
			TargetItemIds: offer.TargetItemIds,
			Price:         offer.Price,
			TargetType:    offer.TargetType,
			Type:          "ORDER",
		}
		appliedSaleRecordCartOffers = append(appliedSaleRecordCartOffers, cartOffer)
	}
	return appliedSaleRecordCartOffers
}

func (itemEvent *OrderEventItem) makeSaleRecordItemOffers(item *Item, tenantCode string) []AppliedSaleRecordItemOffer {
	if item == nil || item.Offer == nil {
		return nil
	}
	if itemEvent.TotalDiscountPrice == 0 {
		return nil
	}
	var appliedSaleRecordItemOffers = make([]AppliedSaleRecordItemOffer, 0)
	saleRecordItemOffer := AppliedSaleRecordItemOffer{
		TenantCode:       tenantCode,
		OfferNo:          item.Offer.No,
		ItemCode:         itemEvent.ItemCode,
		ItemCodes:        itemEvent.ItemCode,
		Price:            item.Offer.DiscountPrice,
		Type:             "ORDER",
		TransactionDtlId: itemEvent.Id,
	}
	appliedSaleRecordItemOffers = append(appliedSaleRecordItemOffers, saleRecordItemOffer)
	return appliedSaleRecordItemOffers
}
func (itemEvent *OrderEventItem) makeSaleRecordItemCartOffers() []SaleRecordItemAppliedCartOffer {
	var appliedSaleRecordCartOffers = make([]SaleRecordItemAppliedCartOffer, 0)
	for _, offer := range itemEvent.GroupOffers {
		saleRecordItemOffer := SaleRecordItemAppliedCartOffer{
			OfferNo:    offer.OfferNo,
			CouponNo:   offer.CouponNo,
			TargetType: offer.TargetType,
			IsTarget:   offer.IsTarget,
			Price:      offer.Price,
			Type:       "ORDER",
		}
		appliedSaleRecordCartOffers = append(appliedSaleRecordCartOffers, saleRecordItemOffer)
	}
	return appliedSaleRecordCartOffers
}
func (orderEvent *OrderEvent) makePayments(ctx context.Context) ([]AssortedSaleRecordPayment, error) {
	pays, err := (Pay{}).GetPays(ctx, orderEvent.Id, 0)
	if err != nil {
		return nil, err
	}
	return Pay{}.MakePostPayments(pays), nil
}

func (orderEvent *OrderEvent) getDiscountPrice(setting *number.Setting) (float64, float64) {
	if len(orderEvent.Offers) == 0 {
		return 0, 0
	}
	var couponDiscountPrice, offerDiscountPrice float64
	for _, offer := range orderEvent.Offers {
		if offer.CouponNo != "" {
			couponDiscountPrice = number.ToFixed(couponDiscountPrice+offer.Price, nil)
		} else {
			offerDiscountPrice = number.ToFixed(offerDiscountPrice+offer.Price, nil)
		}
	}
	return offerDiscountPrice, couponDiscountPrice
}

func (orderEvent *OrderEvent) makeEmpId(ctx context.Context) (string, error) {
	if orderEvent.Offers == nil || orderEvent.CustomerId <= 0 {
		return "", nil
	}

	for _, offer := range orderEvent.Offers {
		if offer.CouponNo == "" {
			continue
		}
		coupon, err := GetCoupon(ctx, offer.CouponNo)
		if err != nil {
			return "", err
		}
		if !coupon.IsInternal {
			continue
		}
		member, err := GetMember(ctx, orderEvent.CustomerId, orderEvent.TenantCode)
		if err != nil {
			return "", err
		}
		if member == nil {
			continue
		}
		return member.HrEmpNo, nil
	}
	return "", nil
}

func (orderEvent *OrderEvent) distributeCartOfferPrice(setting *number.Setting) error {
	if len(orderEvent.Offers) == 0 {
		return nil
	}

	getItemIds := func(cartOffer OfferEvent) []string {
		itemIds := make([]string, 0)
		if cartOffer.TargetItemIds != "" {
			itemIds = strings.Split(cartOffer.TargetItemIds, ",")
		} else {
			itemIds = strings.Split(cartOffer.ItemIds, ",")
		}
		return itemIds
	}

	for _, cartOffer := range orderEvent.Offers {
		if cartOffer.Price == 0 {
			continue
		}
		itemIds := getItemIds(cartOffer)
		if cartOffer.TargetType == "gift" {
			for i, orderItem := range orderEvent.Items {
				for _, itemId := range itemIds {
					if itemId != strconv.FormatInt(orderItem.Id, 10) {
						continue
					}
					orderEvent.Items[i].TotalDistributedCartOfferPrice = orderItem.TotalListPrice
				}
			}
		} else {
			orderEvent.calculateDistributePrice(cartOffer.Price, itemIds, setting)
		}

	}
	return nil
}

func (orderEvent *OrderEvent) calculateDistributePrice(price float64, itemIds []string, setting *number.Setting) {
	distributeData := orderEvent.makeDistributeData(itemIds, setting)
	CalculateDistributeAmt(distributeData, price, setting)

	for i, orderItem := range orderEvent.Items {
		for _, distributeItem := range distributeData.DistributeItems {
			if distributeItem.Id != orderItem.Id {
				continue
			}
			orderEvent.Items[i].TotalDistributedCartOfferPrice = number.ToFixed(orderEvent.Items[i].TotalDistributedCartOfferPrice+distributeItem.DistributeAmt, nil)
		}
	}
}

func (orderEvent *OrderEvent) makeDistributeData(itemIds []string, setting *number.Setting) *DistributeData {
	createDistributeData := func(orderItem OrderEventItem, distributeData *DistributeData) {
		distributeData.DistributeItemsTotalAmt = number.ToFixed(distributeData.DistributeItemsTotalAmt+orderItem.TotalSalePrice, setting)
		distributeData.DistributeItems = append(distributeData.DistributeItems, DistributeItem{
			Id:                 orderItem.Id,
			ItemCode:           orderItem.ItemCode,
			DistributeAmt:      0,
			ItemTotalSalePrice: orderItem.TotalSalePrice,
		})
	}
	distributeData := DistributeData{}
	for _, orderItem := range orderEvent.Items {
		for _, itemId := range itemIds {
			if strconv.FormatInt(orderItem.Id, 10) != itemId {
				continue
			}
			createDistributeData(orderItem, &distributeData)
		}
	}

	return &distributeData
}
