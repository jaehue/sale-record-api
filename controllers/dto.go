package controllers

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/hublabs/sale-record-api/models"

	"github.com/pangpanglabs/goutils/number"
)

type SaleRecordInput struct {
	OrderId              int64                 `json:"orderId" validate:"gte=0"`
	RefundId             int64                 `json:"refundId" validate:"gte=0"`
	ChannelId            int64                 `json:"channelId"`
	StoreId              int64                 `json:"storeId" validate:"gte=0"`
	OuterOrderNo         string                `json:"outerOrderNo"`
	TenantCode           string                `json:"tenantCode" validate:"required"`
	TotalListPrice       float64               `json:"totalListPrice"  validate:"required"`
	TotalDiscountPrice   float64               `json:"totalDiscountPrice" validate:"gte=0"`
	TotalPaymentPrice    float64               `json:"totalPaymentPrice"  validate:"required"`
	FreightPrice         float64               `json:"freightPrice" `
	ChannelType          string                `json:"channelType"  validate:"required"`
	CreatedId            int64                 `json:"createdId"  validate:"gte=0"`
	SalesmanId           int64                 `json:"salesmanId" validate:"gte=0"`
	TransactionType      string                `json:"transactionType"  validate:"required"`
	OfflineShopCode      string                `json:"offlineShopCode" `
	SaleRecordDtlInputs  []SaleRecordDtlInput  `json:"saleRecordDtlInputs" validate:"required,dive,required"`
	SaleRecordCartOffers []SaleRecordCartOffer `json:"saleRecordCartOffers"`
	SaleRecordPayments   []SaleRecordPayment   `json:"saleRecordPayments" validate:"required,dive,required"`
}

type SaleRecordDtlInput struct {
	OrderItemId        int64   `json:"orderItemId" validate:"gte=0"`
	RefundItemId       int64   `json:"refundItemId" validate:"gte=0"`
	BrandId            int64   `json:"brandId" validate:"required"`
	BrandCode          string  `json:"brandCode" validate:"required"`
	ProductId          int64   `json:"productId" validate:"required"`
	SkuId              int64   `json:"skuId" validate:"required"`
	FeeRate            float64 `json:"feeRate" validate:""`
	ListPrice          float64 `json:"listPrice" validate:"required"`
	SalePrice          float64 `json:"salePrice" validate:"required"`
	Quantity           int     `json:"quantity" validate:"required"`
	TotalListPrice     float64 `json:"totalListPrice" validate:"required"`
	TotalDiscountPrice float64 `json:"totalDiscountPrice" validate:"gte=0"`
	TotalPaymentPrice  float64 `json:"totalPaymentPrice" validate:"required"`
}

type SaleRecordCartOffer struct {
	OfferNo      string  `json:"offerNo" validate:"required"`
	OrderItemIds string  `json:"orderItemIds"`
	Price        float64 `json:"price" validate:"gte=0"`
}

type SaleRecordPayment struct {
	PayAmt    float64 `json:"payAmt" validate:"required"`
	PayMethod string  `json:"payMethod" validate:"required"`
}

func (input SaleRecordInput) Validate(ctx context.Context) error {
	var totalPaymentPrice float64
	var totalDiscountPrice float64
	var totalListPrice float64
	for _, dtlInput := range input.SaleRecordDtlInputs {
		totalPaymentPrice = number.ToFixed(totalPaymentPrice+dtlInput.TotalPaymentPrice, nil)
		totalDiscountPrice = number.ToFixed(totalDiscountPrice+dtlInput.TotalDiscountPrice, nil)
		totalListPrice = number.ToFixed(totalListPrice+dtlInput.TotalListPrice, nil)
	}

	if totalPaymentPrice != input.TotalPaymentPrice || totalDiscountPrice != input.TotalDiscountPrice || totalListPrice != input.TotalListPrice {
		return errors.New(string(models.TotalPriceError))
	}
	if input.CreatedId == 0 {
		return errors.New(string(models.CreatedIdError))
	}
	if totalDiscountPrice != number.ToFixed(totalListPrice-totalPaymentPrice, nil) {
		return errors.New(string(models.DiscountPriceError))
	}
	if totalDiscountPrice > 0 {
		if len(input.SaleRecordCartOffers) == 0 {
			return errors.New(string(models.DiscountPriceNotMatchOfferError))
		}
		var totalCartOfferPrice float64
		for _, cartOffer := range input.SaleRecordCartOffers {
			if cartOffer.OfferNo == "" {
				return errors.New(string(models.OfferNoError))
			}
			totalCartOfferPrice = number.ToFixed(totalCartOfferPrice+cartOffer.Price, nil)
		}
		if totalCartOfferPrice != totalDiscountPrice {
			return errors.New(string(models.DiscountPriceNotMatchOfferError))
		}
	}

	isBrandExist := func(brandId int64, brandCode string, brands []models.StoreBrand) bool {
		for _, storeBrand := range brands {
			if brandCode == storeBrand.Code && brandId == storeBrand.Id {
				return true
			}
		}
		return false
	}
	if input.StoreId == 0 {
		return errors.New(string(models.StoreIdError))
	}
	store, err := models.GetStore(ctx, input.StoreId)
	if err != nil {
		return err
	}
	if store == nil {
		return errors.New(string(models.StoreNotExistError))
	}
	for _, dtlInput := range input.SaleRecordDtlInputs {
		if !isBrandExist(dtlInput.BrandId, dtlInput.BrandCode, store.Brands) {
			return errors.New(string(models.BrandNotMatchError))
		}
	}

	for _, dtlInput := range input.SaleRecordDtlInputs {
		sku, err := models.GetSkuBy(ctx, dtlInput.SkuId)
		if err != nil {
			return err
		}
		if sku == nil {
			return errors.New(string(models.SkuNotExistError))
		}

		if sku.Product.ListPrice != dtlInput.ListPrice {
			return errors.New(string(models.SkuListPriceError))
		}

		if sku.Product.Id != dtlInput.ProductId {
			return errors.New(string(models.ProductNotMatchError))
		}

		if sku.Product.Brand.Code != dtlInput.BrandCode {
			return errors.New(string(models.ProductBrandError))
		}
	}

	return nil
}

func (input SaleRecordInput) MakeSaleRecord() models.AssortedSaleRecord {
	var saleRecord = models.AssortedSaleRecord{}
	saleRecord.OrderId = input.OrderId
	saleRecord.RefundId = input.RefundId
	saleRecord.TenantCode = input.TenantCode
	saleRecord.CashPrice = input.TotalPaymentPrice
	saleRecord.Created = time.Now().UTC()
	saleRecord.Modified = time.Now().UTC()
	if input.ChannelType == "EMALL" {
		saleRecord.CreatedBy = "excel-upload"
		saleRecord.ModifiedBy = "excel-upload"
	} else if input.ChannelType == "TMALL" {
		saleRecord.CreatedBy = "sale-record-tmall"
		saleRecord.ModifiedBy = "sale-record-tmall"
	} else {
		saleRecord.CreatedBy = strconv.FormatInt(input.CreatedId, 10)
		saleRecord.ModifiedBy = saleRecord.CreatedBy
		saleRecord.SalesmanId = input.SalesmanId
	}
	if input.RefundId == 0 {
		saleRecord.TransactionStatus = "BuyerReceivedConfirmed"
	} else {
		saleRecord.IsRefund = true
		saleRecord.TransactionStatus = "RefundOrderSuccess"
	}
	if input.OfflineShopCode != "" {
		salesman := strings.Split(strings.Replace(input.OfflineShopCode, "，", ",", 1), ",")
		saleRecord.SalesmanShopCode = salesman[0]
		if len(salesman) > 1 {
			saleRecord.SalesmanEmpId = salesman[len(salesman)-1]
		}
	}
	saleRecord.DiscountOfferPrice = input.TotalDiscountPrice
	saleRecord.TotalTransactionPrice = input.TotalPaymentPrice
	saleRecord.TransactionChannelType = input.ChannelType
	saleRecord.TransactionCreateDate = time.Now().UTC()
	saleRecord.TransactionType = input.TransactionType
	saleRecord.TransactionUpdateDate = time.Now().UTC()
	saleRecord.IsOutPaid = true
	saleRecord.OuterOrderNo = input.OuterOrderNo
	saleRecord.StoreId = input.StoreId
	saleRecord.TotalDiscountPrice = input.TotalDiscountPrice
	saleRecord.TotalListPrice = input.TotalListPrice
	saleRecord.TotalSalePrice = input.TotalPaymentPrice
	saleRecord.FreightPrice = input.FreightPrice
	saleRecord.TransactionCreatedId = input.CreatedId
	saleRecord.AppliedSaleRecordCartOffers = input.MakeSaleRecordCartOffers()
	saleRecord.SaleRecordDtls = input.MakeSaleRecordDtls(saleRecord)
	saleRecord.Payments = input.MakeSaleRecordPayments()
	return saleRecord
}

func (input SaleRecordInput) MakeSaleRecordDtls(saleRecord models.AssortedSaleRecord) []models.AssortedSaleRecordDtl {
	var saleRecordDtls = make([]models.AssortedSaleRecordDtl, 0)
	for _, dtlInput := range input.SaleRecordDtlInputs {
		saleRecordDtl := models.AssortedSaleRecordDtl{
			BrandCode:                      dtlInput.BrandCode,
			BrandId:                        dtlInput.BrandId,
			Created:                        time.Now().UTC(),
			CreatedBy:                      saleRecord.CreatedBy,
			Modified:                       time.Now().UTC(),
			ModifiedBy:                     saleRecord.CreatedBy,
			DistributedCashPrice:           dtlInput.TotalPaymentPrice,
			TotalDistributedCartOfferPrice: dtlInput.TotalDiscountPrice,
			TotalDistributedPaymentPrice:   dtlInput.TotalPaymentPrice,
			IsDelivery:                     false,
			FeeRate:                        dtlInput.FeeRate,
			ListPrice:                      dtlInput.ListPrice,
			OrderItemId:                    dtlInput.OrderItemId,
			RefundItemId:                   dtlInput.RefundItemId,
			ProductId:                      dtlInput.ProductId,
			Quantity:                       dtlInput.Quantity,
			SalePrice:                      dtlInput.SalePrice,
			SkuId:                          dtlInput.SkuId,
			Status:                         saleRecord.TransactionStatus,
			TotalDiscountPrice:             dtlInput.TotalDiscountPrice,
			TotalListPrice:                 dtlInput.TotalListPrice,
			TotalSalePrice:                 dtlInput.TotalPaymentPrice,
			TotalTransactionPrice:          dtlInput.TotalPaymentPrice,
		}
		saleRecordDtls = append(saleRecordDtls, saleRecordDtl)
	}
	return saleRecordDtls
}

func (input SaleRecordInput) MakeSaleRecordCartOffers() []models.AppliedSaleRecordCartOffer {
	var appliedSaleRecordCartOffers = make([]models.AppliedSaleRecordCartOffer, 0)
	for _, offer := range input.SaleRecordCartOffers {
		var cartOffer = models.AppliedSaleRecordCartOffer{
			TenantCode: input.TenantCode,
			OfferNo:    offer.OfferNo,
			ItemIds:    offer.OrderItemIds,
			Price:      offer.Price,
		}
		appliedSaleRecordCartOffers = append(appliedSaleRecordCartOffers, cartOffer)
	}
	return appliedSaleRecordCartOffers
}

func (input SaleRecordInput) MakeSaleRecordPayments() []models.AssortedSaleRecordPayment {
	var payments = make([]models.AssortedSaleRecordPayment, 0)
	paymentEvent := models.AssortedSaleRecordPayment{
		SeqNo:     1,
		PayMethod: "CASH",
		PayAmt:    input.TotalPaymentPrice,
		CreatedAt: time.Now().UTC(),
	}
	payments = append(payments, paymentEvent)
	return payments
}
