package adapters

import (
	"github.com/hublabs/common/eventconsume"
	"github.com/hublabs/sale-record-api/models"

	"github.com/pangpanglabs/goutils/kafka"
	"github.com/sirupsen/logrus"
)

func OrderEventConsumer(serviceName string, orderConfit kafka.Config, filters ...eventconsume.Filter) error {
	return eventconsume.NewEventConsumer(serviceName, orderConfit.Brokers, orderConfit.Topic, filters).Handle(handleOrderEvent)
}

func handleOrderEvent(c eventconsume.ConsumeContext) error {
	var orderEvent models.Event
	if err := c.Bind(&orderEvent); err != nil {
		logrus.WithField("Error", err).Info("OrderEvent bind error!")
		return err
	}

	ctx := c.Context()
	if err := (models.OrderEventHandler{}).Handle(ctx, orderEvent); err != nil {
		if orderEvent.Payload.Refunds == nil {
			if saveErr := models.SaveRefundEventError(ctx, err.Error(), orderEvent); saveErr != nil {
				return err
			}
		} else {
			if saveErr := models.SaveOrderEventError(ctx, err.Error(), orderEvent); saveErr != nil {
				return err
			}
			return err
		}
	}
	return nil
}
