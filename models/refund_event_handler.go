package models

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/pangpanglabs/goutils/number"
)

type RefundEvent struct {
	Id                 int64             `json:"id"`
	OuterOrderNo       string            `json:"outerOrderNo"`
	TenantCode         string            `json:"tenantCode"`
	StoreId            int64             `json:"storeId" `
	ChannelId          int64             `json:"channelId" `
	RefundType         string            `json:"refundType"`
	CustomerId         int64             `json:"customerId"`
	SalesmanId         int64             `json:"salesmanId" `
	TotalListPrice     float64           `json:"totalListPrice"`
	TotalSalePrice     float64           `json:"totalSalePrice"`
	TotalDiscountPrice float64           `json:"totalDiscountPrice"`
	FreightPrice       float64           `json:"freightPrice"`
	TotalRefundPrice   float64           `json:"totalRefundPrice"`
	Mileage            float64           `json:"mileage"`
	MileagePrice       float64           `json:"mileagePrice"`
	ObtainMileage      float64           `json:"obtainMileage"`
	CashPrice          float64           `json:"cashPrice"`
	Status             string            `json:"status"`
	IsOutPaid          bool              `json:"isOutPaid" `
	CustRemark         string            `json:"custRemark" `
	Items              []RefundEventItem `json:"items,omitempty"`
	RefundReason       string            `json:"refundReason"`
	RefuseReason       string            `json:"refuseReason"`
	Offers             []OfferEvent      `json:"offers,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
	CreatedId          int64             `json:"createdId"`
}
type RefundEventItem struct {
	Id                             int64                 `json:"id"`
	OuterOrderItemNo               string                `json:"outerOrderItemNo"`
	OrderItemId                    int64                 `json:"orderItemId"`
	SeparateId                     int64                 `json:"separateId"`
	StockDistributionItemId        int64                 `json:"stockDistributionItemId"`
	ItemCode                       string                `json:"itemCode"`
	ItemName                       string                `json:"itemName"`
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
	TotalRefundPrice               float64               `json:"totalRefundPrice"`
	Mileage                        float64               `json:"mileage"`
	MileagePrice                   float64               `json:"mileagePrice"`
	ObtainMileage                  float64               `json:"obtainMileage"`
	CashPrice                      float64               `json:"cashPrice"`
	TotalDistributedCartOfferPrice float64               `json:"totalDistributedCartOfferPrice"`
	ItemFee                        float64               `json:"itemFee"`
	FeeRate                        float64               `json:"feeRate"`
	Status                         string                `json:"status"`
	IsDelivery                     bool                  `json:"isDelivery" `
	GroupOffers                    []OrderItemOfferEvent `json:"groupOffers,omitempty"`
	CreatedAt                      time.Time             `json:"createdAt"`
	UpdatedAt                      time.Time             `json:"updatedAt"`
}

func (RefundEvent) HandleRefundEvent(ctx context.Context, event Event) (bool, *AssortedSaleRecord, error) {
	refundEvent := event.Payload.Refunds[0]
	if exists, saleRecord, err := CheckSaleRecordExits(ctx, 0, refundEvent.Id, refundEvent.RefundType, ""); err != nil {
		return false, nil, err
	} else if exists {
		return true, saleRecord, nil
	}
	if err := refundEvent.RefundValidate(ctx); err != nil {
		return false, nil, err
	}
	numberSetting, baseTrimCode, err := GetRoundSetting(ctx, refundEvent.StoreId)
	if err != nil {
		return false, nil, err
	}
	saleRecord, err := refundEvent.makeAssortedSaleRecord(ctx, numberSetting)
	if err != nil {
		return false, nil, err
	}
	saleRecord.BaseTrimCode = baseTrimCode
	saleRecord.OrderId = event.Payload.Id

	return false, saleRecord, nil
}

func (refundEvent *RefundEvent) makeAssortedSaleRecord(ctx context.Context, setting *number.Setting) (*AssortedSaleRecord, error) {
	// refundEvent.distributeCartOfferPrice(setting)
	var saleRecode AssortedSaleRecord
	saleRecode.TenantCode = refundEvent.TenantCode
	saleRecode.CashPrice = refundEvent.CashPrice
	saleRecode.ChannelId = refundEvent.ChannelId
	saleRecode.Created = time.Now().UTC()
	saleRecode.CreatedBy = strconv.FormatInt(refundEvent.SalesmanId, 10)
	saleRecode.Modified = time.Now().UTC()
	saleRecode.ModifiedBy = strconv.FormatInt(refundEvent.SalesmanId, 10)
	saleRecode.CustomerId = refundEvent.CustomerId
	saleRecode.RefundId = refundEvent.Id
	saleRecode.IsRefund = true
	saleRecode.DiscountOfferPrice, saleRecode.DiscountCouponPrice = refundEvent.getDiscountPrice(setting)
	saleRecode.TotalTransactionPrice = refundEvent.TotalRefundPrice
	saleRecode.TransactionChannelType = refundEvent.RefundType
	saleRecode.TransactionCreateDate = refundEvent.CreatedAt
	saleRecode.TransactionStatus = refundEvent.Status
	saleRecode.TransactionType = "MINUS"
	saleRecode.TransactionUpdateDate = refundEvent.CreatedAt
	saleRecode.FreightPrice = refundEvent.FreightPrice
	saleRecode.IsOutPaid = refundEvent.IsOutPaid
	saleRecode.Mileage = refundEvent.Mileage
	saleRecode.MileagePrice = refundEvent.MileagePrice
	saleRecode.ObtainMileage = refundEvent.ObtainMileage
	saleRecode.OuterOrderNo = refundEvent.OuterOrderNo
	saleRecode.SalesmanId = refundEvent.SalesmanId
	saleRecode.StoreId = refundEvent.StoreId
	saleRecode.TenantCode = refundEvent.TenantCode
	saleRecode.TotalDiscountPrice = refundEvent.TotalDiscountPrice
	saleRecode.TotalListPrice = refundEvent.TotalListPrice
	saleRecode.TotalSalePrice = refundEvent.TotalSalePrice
	saleRecode.TransactionCreatedId = refundEvent.CreatedId

	var err error
	if saleRecode.SaleRecordDtls, err = refundEvent.makeAssortedSaleRecordDtls(ctx, setting); err != nil {
		return nil, err
	}
	if saleRecode.Payments, err = refundEvent.makePayments(ctx); err != nil {
		return nil, err
	}
	saleRecode.AppliedSaleRecordCartOffers = refundEvent.makeSaleRecordCartOffers()
	return &saleRecode, nil
}

func (refundEvent RefundEvent) makeAssortedSaleRecordDtls(ctx context.Context, setting *number.Setting) ([]AssortedSaleRecordDtl, error) {
	var saleRecodeDtls = make([]AssortedSaleRecordDtl, 0)
	for _, refundItemEvent := range refundEvent.Items {
		item, err := Item{}.GetItemByCode(ctx, refundItemEvent.ItemCode)
		if err != nil {
			return nil, err
		}
		saleRecordDtl := AssortedSaleRecordDtl{
			BrandCode:                      item.Sku.Product.Brand.Code,
			BrandId:                        item.Sku.Product.Brand.Id,
			Created:                        refundItemEvent.CreatedAt,
			CreatedBy:                      "kafka-listener",
			Modified:                       refundItemEvent.UpdatedAt,
			ModifiedBy:                     "kafka-listener",
			IsDelivery:                     refundItemEvent.IsDelivery,
			ItemCode:                       refundItemEvent.ItemCode,
			ItemName:                       refundItemEvent.ItemName,
			ListPrice:                      item.Sku.Product.ListPrice,
			ProductId:                      item.Sku.Product.Id,
			Quantity:                       refundItemEvent.Quantity,
			Mileage:                        refundItemEvent.Mileage,
			MileagePrice:                   refundItemEvent.MileagePrice,
			ObtainMileage:                  refundItemEvent.ObtainMileage,
			RefundItemId:                   refundItemEvent.Id,
			OrderItemId:                    refundItemEvent.OrderItemId,
			SalePrice:                      item.SalePrice,
			SkuId:                          refundItemEvent.SkuId,
			SkuImg:                         refundItemEvent.SkuImg,
			Status:                         refundItemEvent.Status,
			TotalDiscountPrice:             refundItemEvent.TotalDiscountPrice,
			TotalListPrice:                 refundItemEvent.TotalListPrice,
			TotalSalePrice:                 refundItemEvent.TotalSalePrice,
			TotalTransactionPrice:          refundItemEvent.TotalRefundPrice,
			FeeRate:                        refundItemEvent.FeeRate,
			DistributedCashPrice:           number.ToFixed(refundItemEvent.TotalRefundPrice-refundItemEvent.TotalDistributedCartOfferPrice-refundItemEvent.MileagePrice, nil),
			TotalDistributedCartOfferPrice: refundItemEvent.TotalDistributedCartOfferPrice,
			TotalDistributedItemOfferPrice: number.ToFixed(refundItemEvent.TotalListPrice-refundItemEvent.TotalSalePrice, nil),
			TotalDistributedPaymentPrice:   number.ToFixed(refundItemEvent.TotalRefundPrice-refundItemEvent.TotalDistributedCartOfferPrice, nil),
			AppliedSaleRecordItemOffers:    refundItemEvent.makeSaleRecordItemOffers(item, refundEvent.TenantCode),
			AppliedSaleRecordCartOffers:    refundItemEvent.makeSaleRecordItemCartOffers(),
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

func (refundEvent *RefundEvent) makeSaleRecordCartOffers() []AppliedSaleRecordCartOffer {
	var appliedSaleRecordCartOffers = make([]AppliedSaleRecordCartOffer, 0)
	for _, offer := range refundEvent.Offers {
		var cartOffer = AppliedSaleRecordCartOffer{
			TenantCode:    refundEvent.TenantCode,
			OfferNo:       offer.OfferNo,
			CouponNo:      offer.CouponNo,
			ItemIds:       offer.ItemIds,
			TargetItemIds: offer.TargetItemIds,
			Price:         offer.Price,
			Type:          "REFUND",
		}
		appliedSaleRecordCartOffers = append(appliedSaleRecordCartOffers, cartOffer)
	}
	return appliedSaleRecordCartOffers
}

func (itemEvent *RefundEventItem) makeSaleRecordItemOffers(item *Item, tenantCode string) []AppliedSaleRecordItemOffer {
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
		Type:             "REFUND",
		TransactionDtlId: itemEvent.Id,
	}
	appliedSaleRecordItemOffers = append(appliedSaleRecordItemOffers, saleRecordItemOffer)
	return appliedSaleRecordItemOffers
}
func (itemEvent *RefundEventItem) makeSaleRecordItemCartOffers() []SaleRecordItemAppliedCartOffer {
	var appliedSaleRecordCartOffers = make([]SaleRecordItemAppliedCartOffer, 0)
	for _, offer := range itemEvent.GroupOffers {
		saleRecordItemOffer := SaleRecordItemAppliedCartOffer{
			OfferNo:    offer.OfferNo,
			CouponNo:   offer.CouponNo,
			TargetType: offer.TargetType,
			IsTarget:   offer.IsTarget,
			Price:      offer.Price,
			Type:       "REFUND",
		}
		appliedSaleRecordCartOffers = append(appliedSaleRecordCartOffers, saleRecordItemOffer)
	}
	return appliedSaleRecordCartOffers
}
func (refundEvent *RefundEvent) getDiscountPrice(setting *number.Setting) (float64, float64) {
	if len(refundEvent.Offers) == 0 {
		return 0, 0
	}
	var couponDiscountPrice, offerDiscountPrice float64
	for _, offer := range refundEvent.Offers {
		if offer.CouponNo != "" {
			couponDiscountPrice = number.ToFixed(couponDiscountPrice+offer.Price, nil)
		} else {
			offerDiscountPrice = number.ToFixed(offerDiscountPrice+offer.Price, nil)
		}
	}
	return offerDiscountPrice, couponDiscountPrice
}

func (refundEvent *RefundEvent) makePayments(ctx context.Context) ([]AssortedSaleRecordPayment, error) {
	pays, err := (Pay{}).GetPays(ctx, 0, refundEvent.Id)
	if err != nil {
		return nil, err
	}
	return Pay{}.MakePostPayments(pays), nil
}

func (refundEvent *RefundEvent) distributeCartOfferPrice(setting *number.Setting) error {
	if len(refundEvent.Offers) == 0 {
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

	for _, cartOffer := range refundEvent.Offers {
		if cartOffer.Price == 0 {
			continue
		}
		itemIds := getItemIds(cartOffer)
		if cartOffer.TargetType == "gift" {
			for i, refundItem := range refundEvent.Items {
				for _, itemId := range itemIds {
					if itemId != strconv.FormatInt(refundItem.OrderItemId, 10) {
						continue
					}
					refundEvent.Items[i].TotalDistributedCartOfferPrice = refundItem.TotalListPrice
				}
			}
		} else {
			refundEvent.calculateDistributePrice(cartOffer.Price, itemIds, setting)
		}
	}
	return nil
}

func (refundEvent *RefundEvent) calculateDistributePrice(price float64, itemIds []string, setting *number.Setting) {
	distributeData := refundEvent.makeDistributeData(itemIds, setting)
	CalculateDistributeAmt(distributeData, price, setting)

	for i, refundItem := range refundEvent.Items {
		for _, distributeItem := range distributeData.DistributeItems {
			if distributeItem.Id != refundItem.OrderItemId {
				continue
			}
			refundEvent.Items[i].TotalDistributedCartOfferPrice = number.ToFixed(refundEvent.Items[i].TotalDistributedCartOfferPrice+distributeItem.DistributeAmt, nil)
		}
	}
}

func (refundEvent *RefundEvent) makeDistributeData(itemIds []string, setting *number.Setting) *DistributeData {
	createDistributeData := func(refundItem RefundEventItem, distributeData *DistributeData) {
		distributeData.DistributeItemsTotalAmt = number.ToFixed(distributeData.DistributeItemsTotalAmt+refundItem.TotalSalePrice, setting)
		distributeData.DistributeItems = append(distributeData.DistributeItems, DistributeItem{
			Id:                 refundItem.OrderItemId,
			ItemCode:           refundItem.ItemCode,
			DistributeAmt:      0,
			ItemTotalSalePrice: refundItem.TotalSalePrice,
		})
	}
	distributeData := DistributeData{}
	for _, refundItem := range refundEvent.Items {
		for _, itemId := range itemIds {
			if strconv.FormatInt(refundItem.OrderItemId, 10) != itemId {
				continue
			}
			createDistributeData(refundItem, &distributeData)
		}
	}

	return &distributeData
}
