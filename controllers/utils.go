package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hublabs/common/api"
	"github.com/hublabs/sale-record-api/models"
	"github.com/pangpanglabs/goutils/behaviorlog"

	"github.com/labstack/echo"
)

func ReturnApiFail(c echo.Context, err error) error {
	if err == nil {
		err = api.ErrorUnknown.New(nil)
	}

	behaviorlog.FromCtx(c.Request().Context()).WithError(err)
	var apiError *api.Error
	if ok := errors.As(err, &apiError); ok {
		return c.JSON(apiError.Status(), api.Result{
			Success: false,
			Error:   *apiError,
		})
	}
	return err
}
func DateTimeValidate(startDate, endDate string, term int) error {
	timeLayout := "2006-01-02"

	if startDate == "" && endDate == "" {
		return nil
	}
	if startDate == "" && endDate != "" {
		return errors.New("结束日期为空")
	}
	if startDate != "" && endDate == "" {
		return errors.New("开始日期为空")
	}

	startTime, err := time.Parse(timeLayout, startDate)
	if err != nil {
		return err
	}
	endTime, err := time.Parse(timeLayout, endDate)
	if err != nil {
		return err
	}

	if startTime.After(endTime) {
		return err
	}
	if startTime.AddDate(0, 0, term-1).Before(endTime) {
		return errors.New(fmt.Sprintf("查询期间不能大于%d天", term))
	}
	return nil
}

func DateParseToUtc(date string) (timeUtc time.Time, err error) {
	timeLayout := "2006-01-02"
	timeLoc, err := time.Parse(timeLayout, date)
	if err != nil {
		return
	}
	timeUtc = timeLoc.Add(time.Hour * -8)
	return
}

func renderFail(c echo.Context, saleRecord *models.AssortedSaleRecord, saleRecordInput SaleRecordInput, errStr string) error {
	errType, errMsg, detail := models.ErrorTypeMessage(errStr)
	if saveErr := models.SaveSaleRecordSuccess(c.Request().Context(), saleRecord, false, errType, saleRecordInput); saveErr != nil {
		errMsg = saveErr.Error()
	}
	return c.JSON(http.StatusBadRequest, api.Result{
		Success: false,
		Error: api.Error{
			Message: detail,
			Details: errMsg,
		},
	})
}
