package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/hublabs/common/api"
	"github.com/hublabs/sale-record-api/factory"
	"github.com/hublabs/sale-record-api/models"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/echoswagger"
)

const maxDay = 31

type SaleRecordController struct{}

func (c SaleRecordController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")
	g.GET("/get-by-orderId/:orderId", c.GetByOrderId).AddParamPath(1, "orderId", "orderId").
		AddParamQuery("", "channelType", "channelType", false)
	g.GET("/get-by-refundId/:refundId", c.GetByRefundId).AddParamPath(1, "refundId", "refundId").
		AddParamQuery("", "channelType", "channelType", false)
	g.GET("/get-by-transactionId/:transactionId", c.GetByTransactionId).AddParamPath(1, "transactionId", "transactionId")
	g.GET("", c.GetSaleRecords).AddParamQueryNested(models.SaleRecordSearchInput{})
	g.POST("", c.Create).AddParamBody(SaleRecordInput{}, "body", "SaleRecord", true)
	g.GET("/republish/:transactionId", c.EventRepublish).AddParamPath(1, "transactionId", "transactionId")

}

func (SaleRecordController) GetByOrderId(c echo.Context) error {
	orderId, _ := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if orderId == 0 {
		return c.JSON(http.StatusBadRequest, api.Result{
			Error: factory.NewError(api.ErrorMissParameter, "orderId"),
		})
	}
	channelType := c.QueryParam("channelType")
	saleRecord, err := models.GetSaleRecordByOrderId(c.Request().Context(), orderId, channelType, "PLUS")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.Result{
			Success: false,
			Error: api.Error{
				Message: err.Error(),
			},
		})
	}
	return c.JSON(http.StatusOK, api.Result{
		Success: true,
		Result:  saleRecord,
	})
}

func (SaleRecordController) GetByTransactionId(c echo.Context) error {
	transactionId, _ := strconv.ParseInt(c.Param("transactionId"), 10, 64)
	if transactionId == 0 {
		return c.JSON(http.StatusBadRequest, api.Result{
			Error: factory.NewError(api.ErrorMissParameter, "transactionId"),
		})
	}
	saleRecord, err := models.GetSaleRecordByTransactionId(c.Request().Context(), transactionId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.Result{
			Success: false,
			Error: api.Error{
				Message: err.Error(),
			},
		})
	}
	return c.JSON(http.StatusOK, api.Result{
		Success: true,
		Result:  saleRecord,
	})
}
func (SaleRecordController) EventRepublish(c echo.Context) error {
	transactionId, _ := strconv.ParseInt(c.Param("transactionId"), 10, 64)
	if transactionId == 0 {
		return c.JSON(http.StatusBadRequest, api.Result{
			Error: factory.NewError(api.ErrorMissParameter, "transactionId"),
		})
	}
	saleRecord, err := models.GetSaleRecordByTransactionId(c.Request().Context(), transactionId)
	if err != nil {
		return ReturnApiFail(c, api.ErrorDB.New(err))
	}
	go models.EventPublish{}.PublishSaleRecordEvent(c.Request().Context(), saleRecord)
	return c.JSON(http.StatusOK, api.Result{
		Success: true,
		Result:  saleRecord,
	})
}
func (SaleRecordController) GetByRefundId(c echo.Context) error {
	refundId, _ := strconv.ParseInt(c.Param("refundId"), 10, 64)
	if refundId == 0 {
		return c.JSON(http.StatusBadRequest, api.Result{
			Error: factory.NewError(api.ErrorMissParameter, "refundId"),
		})
	}
	channelType := c.QueryParam("channelType")
	saleRecord, err := models.GetSaleRecordByRefundId(c.Request().Context(), refundId, channelType, "MINUS")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.Result{
			Success: false,
			Error: api.Error{
				Message: err.Error(),
			},
		})
	}
	return c.JSON(http.StatusOK, api.Result{
		Success: true,
		Result:  saleRecord,
	})
}

func (SaleRecordController) GetSaleRecords(c echo.Context) error {
	var searchInput models.SaleRecordSearchInput
	if err := c.Bind(&searchInput); err != nil {
		return c.JSON(http.StatusBadRequest, api.Result{
			Success: false,
			Error: api.Error{
				Message: err.Error(),
			},
		})
	}

	getSearchDate := func() (time.Time, time.Time) {
		var orderStartAt = searchInput.OrderStartAt
		var orderEndAt = searchInput.OrderEndAt
		if searchInput.TransactionId != 0 || searchInput.OrderId == 0 || searchInput.RefundId == 0 {
			return time.Time{}, time.Time{}
		}
		if err := DateTimeValidate(orderStartAt, orderEndAt, maxDay); err != nil {
			return time.Now().UTC().AddDate(0, 0, -maxDay), time.Now().UTC()
		}
		if orderStartAt == "" && orderEndAt == "" {
			return time.Now().UTC().AddDate(0, 0, -maxDay), time.Now().UTC()
		}

		var orderStartAtTime, orderEndAtTime time.Time
		var err error
		if orderStartAt != "" {
			orderStartAtTime, err = DateParseToUtc(orderStartAt)
		}
		if orderEndAt != "" {
			orderEndAtTime, err = DateParseToUtc(orderEndAt)
		}
		if err != nil {
			orderEndAtTime = time.Now().UTC()
			orderStartAtTime = time.Now().UTC().AddDate(0, 0, -maxDay)
		}
		return orderStartAtTime, orderEndAtTime
	}
	searchInput.OrderStartTime, searchInput.OrderEndTime = getSearchDate()
	saleRecords, err := models.GetSaleRecords(c.Request().Context(), searchInput)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.Result{
			Success: false,
			Error: api.Error{
				Message: err.Error(),
			},
		})
	}

	return c.JSON(http.StatusOK, api.Result{
		Success: true,
		Result:  saleRecords,
	})
}

func (SaleRecordController) Create(c echo.Context) error {
	var saleRecordInput SaleRecordInput
	if err := c.Bind(&saleRecordInput); err != nil {
		return ReturnApiFail(c, api.ErrorParameter.New(err))
	}
	if err := c.Validate(saleRecordInput); err != nil {
		return ReturnApiFail(c, api.ErrorParameter.New(err))
	}
	if saleRecordInput.OrderId == 0 && saleRecordInput.RefundId == 0 {
		return ReturnApiFail(c, factory.NewError(api.ErrorMissParameter, "orderId and refundId can't both 0"))
	}
	if saleRecordInput.ChannelType == "POS" && saleRecordInput.SalesmanId == 0 {
		return ReturnApiFail(c, factory.NewError(api.ErrorMissParameter, "salesmanId can't  0"))
	}
	saleRecord := saleRecordInput.MakeSaleRecord()
	ctx := c.Request().Context()
	if exists, saleRecord, err := models.CheckSaleRecordExits(ctx, saleRecordInput.OrderId, saleRecordInput.RefundId, saleRecordInput.ChannelType, saleRecordInput.TransactionType); err != nil {
		return renderFail(c, saleRecord, saleRecordInput, err.Error())
	} else if exists {
		go models.EventPublish{}.PublishSaleRecordEvent(ctx, saleRecord)
		return c.JSON(http.StatusOK, api.Result{
			Success: true,
			Result:  saleRecord,
		})
	}
	if err := saleRecordInput.Validate(ctx); err != nil {
		return renderFail(c, &saleRecord, saleRecordInput, err.Error())
	}

	if err := saleRecord.CreateSaleRecord(c.Request().Context()); err != nil {
		return renderFail(c, &saleRecord, saleRecordInput, err.Error())
	}
	if saveErr := models.SaveSaleRecordSuccess(ctx, &saleRecord, true, "", saleRecordInput); saveErr != nil {
		return c.JSON(http.StatusBadRequest, api.Result{
			Success: false,
			Error: api.Error{
				Message: saveErr.Error(),
			},
		})
	}
	go models.EventPublish{}.PublishSaleRecordEvent(ctx, &saleRecord)
	return c.JSON(http.StatusOK, api.Result{
		Success: true,
		Result:  saleRecord,
	})
}
