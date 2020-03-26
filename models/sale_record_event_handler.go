package models

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/kafka"
)

type SaleRecordFailEvent struct {
	Id          int64     `json:"id"`
	TenantCode  string    `json:"tenantCode" `
	ChannelType string    `json:"channelType" `
	OrderId     int64     `json:"orderId" `
	RefundId    int64     `json:"refundId" `
	StoreId     int64     `json:"storeId" `
	ErrorType   string    `json:"errorType" `
	Error       string    `json:"error" `
	IsSuccess   bool      `json:"isSuccess" `
	OrderEntity string    `json:"orderEntitys" `
	CreatedAt   time.Time `json:"createdAt" `
	UpdatedAt   time.Time `json:"updatedAt" `
}

type SaleRecordEvent struct {
	TransactionId             int64                `json:"transactionId"`
	AssortedSaleRecordDtlList []SaleRecordDtlEvent `json:"assortedSaleRecordDtlList"`
	TenantCode                string               `json:"tenantCode"`
	StoreId                   int64                `json:"storeId"`
	OrderId                   int64                `json:"orderId"`
	OuterOrderNo              string               `json:"outerOrderNo"`
	RefundId                  int64                `json:"refundId"`
	TransactionType           string               `json:"transactionType"`
	TransactionChannelType    string               `json:"transactionChannelType"`
	TransactionStatus         string               `json:"transactionStatus"`
	TransactionCreateDate     time.Time            `json:"transactionCreateDate"`
	TransactionUpdateDate     time.Time            `json:"transactionUpdateDate"`
	TransactionCreatedId      int64                `json:"transactionCreatedId"`
	CustomerId                int64                `json:"customerId"`
	EmpId                     string               `json:"empId"`
	SalesmanId                int64                `json:"salesmanId"`
	SalesmanEmpId             string               `json:"salesmanEmpId"`
	SalesmanShopCode          string               `json:"salesmanShopCode"`
	TotalPrice                TotalPriceEvent      `json:"totalPrice"`
	FreightPrice              float64              `json:"freightPrice"`
	Mileage                   float64              `json:"mileage"`
	MileagePrice              float64              `json:"mileagePrice"`
	ObtainMileage             float64              `json:"obtainMileage"`
	CashPrice                 float64              `json:"cashPrice"`
	IsOutPaid                 bool                 `json:"isOutPaid"`
	CartOffers                []CartOfferEvent     `json:"cartOffers"`
	Committed                 CommittedEvent       `json:"committed"`
	OrderCreated              time.Time            `json:"orderCreated"`
	BaseTrimCode              string               `json:"baseTrimCode"`
	Payments                  []PaymentEvent       `json:"payments"`
}

type SaleRecordDtlEvent struct {
	Id               int64                 `json:"id"`
	OrderItemId      int64                 `json:"orderItemId"`
	RefundItemId     int64                 `json:"refundItemId"`
	BrandId          int64                 `json:"brandId"`
	BrandCode        string                `json:"brandCode"`
	ItemCode         string                `json:"itemCode"`
	ItemName         string                `json:"itemName"`
	ProductId        int64                 `json:"productId"`
	SkuId            int64                 `json:"skuId"`
	SkuImg           string                `json:"skuImg"`
	ListPrice        float64               `json:"listPrice"`
	SalePrice        float64               `json:"salePrice"`
	Quantity         int                   `json:"quantity"`
	Mileage          float64               `json:"mileage"`
	MileagePrice     float64               `json:"mileagePrice"`
	ObtainMileage    float64               `json:"obtainMileage"`
	CashPrice        float64               `json:"cashPrice"`
	TotalPrice       TotalPriceEvent       `json:"totalPrice"`
	DistributedPrice DistributedPriceEvent `json:"distributedPrice"`
	Status           string                `json:"status"`
	ItemFee          float64               `json:"itemFee"`
	FeeRate          float64               `json:"feeRate"`
	ItemOffers       []ItemOfferEvent      `json:"itemOffers"`
	CartOffers       []ItemCartOfferEvent  `json:"cartOffers"`
	Committed        CommittedEvent        `json:"committed"`
}

type CommittedEvent struct {
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"createdBy"`
	Modified   time.Time `json:"modified"`
	ModifiedBy string    `json:"modifiedBy"`
}

type ItemOfferEvent struct {
	OfferId   int64   `json:"offerId"`
	OfferNo   string  `json:"offerNo"`
	CouponNo  string  `json:"couponNo"`
	ItemCodes string  `json:"itemCodes"`
	ItemCode  string  `json:"itemCode"`
	Price     float64 `json:"price"`
}
type ItemCartOfferEvent struct {
	OfferNo    string  `json:"offerNo"`
	CouponNo   string  `json:"couponNo"`
	TargetType string  `josn:"targetType"`
	IsTarget   bool    `josn:"isTarget"`
	Price      float64 `json:"price"`
	Type       string  `json:"type"`
}

type TotalPriceEvent struct {
	ListPrice        float64 `json:"listPrice"`
	SalePrice        float64 `json:"salePrice"`
	DiscountPrice    float64 `json:"discountPrice"`
	TransactionPrice float64 `json:"transactionPrice"`
}

type CartOfferEvent struct {
	OfferId       int64  `json:"offerId"`
	OfferNo       string `json:"offerNo"`
	CouponNo      string `json:"couponNo"`
	ItemIds       string `json:"itemIds"`
	TargetItemIds string `json:"targetItemIds"`
	// ItemCodes       string  `json:"itemCodes"`
	// TargetItemCodes string  `json:"targetItemCodes"`
	Price float64 `json:"price"`
}

type PaymentEvent struct {
	SeqNo     int64     `json:"seqNo"`
	PayMethod string    `json:"payMethod"`
	PayAmt    float64   `json:"payAmt"`
	CreatedAt time.Time `json:"createdAt"`
}

type DistributedPriceEvent struct {
	TotalDistributedItemOfferPrice float64 `json:"totalDistributedItemOfferPrice"`
	TotalDistributedCartOfferPrice float64 `json:"totalDistributedCartOfferPrice"`
	TotalDistributedPaymentPrice   float64 `json:"totalDistributedPaymentPrice"`
	DistributedCashPrice           float64 `json:"distributedCashPrice"`
}

var EventPublisher *SaleRecordEventPublisher
var SaleFailPublisher *SaleFailEventPublisher

type SaleRecordEventPublisher struct {
	producer *kafka.Producer
}

type SaleFailEventPublisher struct {
	producer *kafka.Producer
}

func CreateSaleRecordEventPublisher(kafkaConfig kafka.Config) (*SaleRecordEventPublisher, error) {
	producer, err := kafka.NewProducer(kafkaConfig.Brokers, kafkaConfig.Topic, func(c *sarama.Config) {
		c.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
		c.Producer.Compression = sarama.CompressionGZIP     // Compress messages
		c.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms
	})

	if err != nil {
		return nil, err
	}

	saleRecordEventPublisher := SaleRecordEventPublisher{
		producer: producer,
	}

	return &saleRecordEventPublisher, nil
}

func (publisher SaleRecordEventPublisher) Close() {
	publisher.producer.Close()
}

func (publisher SaleRecordEventPublisher) Publish(message interface{}) error {
	if err := publisher.producer.Send(message); err != nil {
		return err
	}
	return nil
}

func CreateSaleFailEventPublisher(kafkaConfig kafka.Config) (*SaleFailEventPublisher, error) {
	producer, err := kafka.NewProducer(kafkaConfig.Brokers, kafkaConfig.Topic, func(c *sarama.Config) {
		c.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
		c.Producer.Compression = sarama.CompressionGZIP     // Compress messages
		c.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms
	})

	if err != nil {
		return nil, err
	}

	saleFailEventPublisher := SaleFailEventPublisher{
		producer: producer,
	}

	return &saleFailEventPublisher, nil
}

func (publisher SaleFailEventPublisher) Close() {
	publisher.producer.Close()
}

func (publisher SaleFailEventPublisher) Publish(message interface{}) error {
	if err := publisher.producer.Send(message); err != nil {
		return err
	}
	return nil
}

func (saleFail *SaleRecordSuccess) translateEventMessage(ctx context.Context) map[string]interface{} {
	saleFailEvent := saleFail.MakeSaleFailEvent()

	return map[string]interface{}{
		"requestId": behaviorlog.FromCtx(ctx).RequestID,
		"actionId":  behaviorlog.FromCtx(ctx).ActionID,
		"authToken": behaviorlog.FromCtx(ctx).AuthToken,
		"payload":   saleFailEvent,
	}
}

func (saleFail *SaleRecordSuccess) MakeSaleFailEvent() SaleRecordFailEvent {
	var saleRecordFailEvent SaleRecordFailEvent

	saleRecordFailEvent.Id = saleFail.Id
	saleRecordFailEvent.TenantCode = saleFail.TenantCode
	saleRecordFailEvent.ChannelType = saleFail.ChannelType
	saleRecordFailEvent.OrderId = saleFail.OrderId
	saleRecordFailEvent.RefundId = saleFail.RefundId
	saleRecordFailEvent.StoreId = saleFail.StoreId
	saleRecordFailEvent.ErrorType = saleFail.ErrorType
	saleRecordFailEvent.Error = saleFail.Error
	saleRecordFailEvent.IsSuccess = saleFail.IsSuccess
	//saleRecordFailEvent.OrderEntity = saleFail.OrderEntity
	saleRecordFailEvent.CreatedAt = saleFail.CreatedAt
	saleRecordFailEvent.UpdatedAt = saleFail.UpdatedAt

	return saleRecordFailEvent
}
