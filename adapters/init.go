package adapters

import (
	"github.com/hublabs/common/eventconsume"

	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/sirupsen/logrus"
)

func AddConsumers(serviceName string, kafkaConfig echomiddleware.KafkaConfig, filters ...eventconsume.Filter) error {
	orderErr := OrderEventConsumer(serviceName, kafkaConfig, filters...)
	if orderErr != nil {
		logrus.WithFields(logrus.Fields{
			"err": orderErr,
		}).Error("OrderEentConsumers Error")
	}
	return nil
}
