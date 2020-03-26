package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hublabs/common/api"
	"github.com/hublabs/common/auth"
	"github.com/hublabs/common/eventconsume"
	"github.com/hublabs/sale-record-api/adapters"
	"github.com/hublabs/sale-record-api/config"
	"github.com/hublabs/sale-record-api/controllers"
	"github.com/hublabs/sale-record-api/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pangpanglabs/echoswagger"
	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

func main() {
	config := config.Init(os.Getenv("APP_ENV"))

	db := initDB(config.Database.Driver, config.Database.Connection)

	if err := adapters.AddConsumers(config.ServiceName, config.OrderEvent.Kafka,
		eventconsume.Recover(),
		eventconsume.ContextDB(config.ServiceName, db, config.Database.Logger.Kafka),
	); err != nil {
		log.Fatal(err)
	}

	saleRecordEventPublisher, err := models.CreateSaleRecordEventPublisher(config.EventBroker.Kafka)
	if err != nil {
		log.Fatal("error set up SaleRecordEventPublisher", err)
	} else {
		models.EventPublisher = saleRecordEventPublisher
	}
	defer saleRecordEventPublisher.Close()

	saleErrorEventPublisher, err := models.CreateSaleFailEventPublisher(config.ErrorEventBroker.Kafka)
	if err != nil {
		log.Fatal("error set up SaleErrorEventPublisher", err)
	} else {
		models.SaleFailPublisher = saleErrorEventPublisher
	}
	defer saleErrorEventPublisher.Close()

	e := echo.New()
	e.Validator = &Validator{validator: validator.New()}

	r := echoswagger.New(e, "doc", &echoswagger.Info{
		Title:       "SRX SaleRecord API",
		Description: "This is docs for sale-record-api service",
		Version:     "1.0.0",
	})

	r.AddSecurityAPIKey("Authorization", "JWT token", echoswagger.SecurityInHeader)

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	e.GET("/whoami", func(c echo.Context) error {
		return c.String(http.StatusOK, config.ServiceName)
	})

	controllers.SaleRecordController{}.Init(r.Group("SaleRecord", "/v1/sale-records"))
	controllers.OrderEventController{}.Init(r.Group("OrderEvent", "/v1/order-events"))
	controllers.SaleRecordLogController{}.Init(r.Group("SaleRecordLog", "/v1/sale-record-log"))

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(echomiddleware.BehaviorLogger(config.ServiceName, config.BehaviorLog.Kafka))
	e.Use(echomiddleware.ContextDB(config.ServiceName, db, echomiddleware.KafkaConfig(config.Database.Logger.Kafka)))
	e.Use(auth.UserClaimMiddleware("/ping", "/doc"))

	if config.AppEnv != "production" {
		behaviorlog.SetLogLevel(logrus.InfoLevel)
		logrus.SetLevel(logrus.InfoLevel)
	}

	if err := e.Start(":8000"); err != nil {
		log.Println(err)
	}

}
func initDB(driver, connection string) *xorm.Engine {
	db, err := xorm.NewEngine(driver, connection)
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Minute * 10)
	db.ShowSQL()

	if err := models.Init(db); err != nil {
		log.Fatal(err)
	}

	return db
}

type Validator struct {
	validator *validator.Validate
}

func (cv *Validator) Validate(i interface{}) error {
	err := cv.validator.Struct(i)
	if err == nil {
		return nil
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		msg := make([]string, 0)
		for _, err := range errs {
			msg = append(msg, fmt.Sprintf("%v condition: %v ,value: %v", err.Field(), err.ActualTag(), err.Value()))
		}
		err := api.ErrorInvalidFields.New(nil)
		err.Message = strings.Join(msg, ",")
		return err
	}
	return err
}
