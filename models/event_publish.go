package models

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type EventPublish struct{}

func (EventPublish) PublishSaleRecordEvent(ctx context.Context, saleRecord *AssortedSaleRecord) {
	saleRecordEvent := saleRecord.translateEventMessage(ctx)
	if err := EventPublisher.Publish(saleRecordEvent); err != nil {
		logrus.WithError(err).Error(fmt.Sprintf(`TransactionId:%v - publish saleRecord to event borker `, saleRecord.TransactionId))
	}
}

func (EventPublish) PublishSaleFailEvent(ctx context.Context, saleFail *SaleRecordSuccess) {
	saleFailEvent := saleFail.translateEventMessage(ctx)
	if err := SaleFailPublisher.Publish(saleFailEvent); err != nil {
		logrus.WithError(err).Error(fmt.Sprintf(`orderId:%v, refundId:%v, ErrorType:%s  - publish saleFail to event borker `, saleFail.OrderId, saleFail.RefundId, saleFail.ErrorType))
	}
}
