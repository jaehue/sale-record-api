package controllers

import (
	"net/http"
	"time"

	"github.com/hublabs/common/api"
	"github.com/hublabs/sale-record-api/models"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/echoswagger"
)

const logMaxDay = 31

type SaleRecordLogController struct{}

func (c SaleRecordLogController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")

	g.GET("", c.getSaleRecordLogs).AddParamQueryNested(models.SaleRecordLogSearchInput{})
}

func (SaleRecordLogController) getSaleRecordLogs(c echo.Context) error {
	var searchInput models.SaleRecordLogSearchInput
	if err := c.Bind(&searchInput); err != nil {
		return c.JSON(http.StatusBadRequest, api.Result{
			Success: false,
			Error: api.Error{
				Message: err.Error(),
			},
		})
	}
	getSearchDate := func() (time.Time, time.Time) {
		if err := DateTimeValidate(searchInput.StartAt, searchInput.EndAt, logMaxDay); err != nil {
			return time.Now().UTC().AddDate(0, 0, -logMaxDay), time.Now().UTC()
		}
		if searchInput.StartAt == "" && searchInput.EndAt == "" && searchInput.OrderId == 0 && searchInput.RefundId == 0 && searchInput.ChannelType == "" {
			return time.Now().UTC().AddDate(0, 0, -logMaxDay), time.Now().UTC()
		}
		var startAtTime, endAtTime time.Time
		var err error
		if searchInput.StartAt != "" {
			startAtTime, err = DateParseToUtc(searchInput.StartAt)
		}
		if searchInput.EndAt != "" {
			endAtTime, err = DateParseToUtc(searchInput.EndAt)
		}
		if err != nil {
			endAtTime = time.Now().UTC()
			startAtTime = time.Now().UTC().AddDate(0, 0, -maxDay)
		}
		return startAtTime, endAtTime
	}
	searchInput.StartTime, searchInput.EndTime = getSearchDate()
	result, err := models.SaleRecordSuccess{}.GetSaleRecordLogs(c.Request().Context(), searchInput)
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
		Result:  result,
	})
}
