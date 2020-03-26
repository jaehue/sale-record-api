package config

import (
	"os"

	configutil "github.com/pangpanglabs/goutils/config"

	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/pangpanglabs/goutils/jwtutil"
	"github.com/sirupsen/logrus"
)

var config C

type C struct {
	Database struct {
		Driver     string
		Connection string
		Logger     struct {
			Kafka echomiddleware.KafkaConfig
		}
	}
	OrderEvent struct {
		Kafka echomiddleware.KafkaConfig
	}
	EventBroker struct {
		Kafka echomiddleware.KafkaConfig
	}
	ErrorEventBroker struct {
		Kafka echomiddleware.KafkaConfig
	}
	BehaviorLog struct {
		Kafka echomiddleware.KafkaConfig
	}
	Debug    bool
	Services struct {
		ColleagueApi          string
		ProductApi            string
		MembershipAddr        string
		MembershipBenefitAddr string
		CouponAddr            string
		PlaceManagementAddr   string
		PayamtApi             string
	}
	ServiceName string
	AppEnv      string
}

func Init(appEnv string, options ...func(*C)) C {
	config.AppEnv = appEnv
	if err := configutil.Read(appEnv, &config); err != nil {
		logrus.WithError(err).Warn("Fail to load config file")
	}

	if s := os.Getenv("JWT_SECRET"); s != "" {
		jwtutil.SetJwtSecret(s)
	}

	for _, option := range options {
		option(&config)
	}

	return config
}

func Config() C {
	return config
}
