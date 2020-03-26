package controllers

import (
	"net/http"

	"github.com/hublabs/common/api"
	"github.com/hublabs/sale-record-api/models"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/echoswagger"
	"github.com/sirupsen/logrus"
)

type OrderEventController struct{}

func (c OrderEventController) Init(g echoswagger.ApiGroup) {
	g.POST("", c.HandleEvent)
}

func (OrderEventController) HandleEvent(c echo.Context) error {
	var orderEvent models.Event

	if err := c.Bind(&orderEvent); err != nil {
		logrus.WithField("Error", err).Info("OrderEvent bind error!")
		return err
	}

	ctx := c.Request().Context()
	if err := (models.OrderEventHandler{}).Handle(ctx, orderEvent); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, api.Result{
		Success: true,
		Result:  orderEvent,
	})
}
