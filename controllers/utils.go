package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hublabs/common/api"
	"github.com/hublabs/sale-record-api/models"

	"github.com/labstack/echo"
)

type ApiResult struct {
	Result  interface{} `json:"result"`
	Success bool        `json:"success"`
	Error   ApiError    `json:"error"`
}

type ApiError struct {
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

type ArrayResult struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"totalCount"`
}

var (
	// System Error
	ApiErrorSystem             = ApiError{Code: 10001, Message: "System Error"}
	ApiErrorServiceUnavailable = ApiError{Code: 10002, Message: "Service unavailable"}
	ApiErrorRemoteService      = ApiError{Code: 10003, Message: "Remote service error"}
	ApiErrorIPLimit            = ApiError{Code: 10004, Message: "IP limit"}
	ApiErrorPermissionDenied   = ApiError{Code: 10005, Message: "Permission denied"}
	ApiErrorIllegalRequest     = ApiError{Code: 10006, Message: "Illegal request"}
	ApiErrorHTTPMethod         = ApiError{Code: 10007, Message: "HTTP method is not suported for this request"}
	ApiErrorParameter          = ApiError{Code: 10008, Message: "Parameter error"}
	ApiErrorMissParameter      = ApiError{Code: 10009, Message: "Miss required parameter"}
	ApiErrorDB                 = ApiError{Code: 10010, Message: "DB error, please contact the administator"}
	ApiErrorTokenInvaild       = ApiError{Code: 10011, Message: "Token invaild"}
	ApiErrorMissToken          = ApiError{Code: 10012, Message: "Miss token"}
	ApiErrorVersion            = ApiError{Code: 10013, Message: "API version %s invalid"}
	ApiErrorNotFound           = ApiError{Code: 10014, Message: "Resource not found"}
	// Business Error
	ApiErrorUserNotExists  = ApiError{Code: 20001, Message: "User does not exists"}
	ApiErrorPassword       = ApiError{Code: 20002, Message: "Password error"}
	ApiErrorStatusNotAllow = ApiError{Code: 20003, Message: "Status Not Allow"}
)

func ReturnApiFail(c echo.Context, status int, apiError ApiError, err error, v ...interface{}) error {
	str := ""
	if err != nil {
		str = err.Error()
	}
	return c.JSON(status, ApiResult{
		Success: false,
		Error: ApiError{
			Code:    apiError.Code,
			Message: fmt.Sprintf(apiError.Message, v...),
			Details: str,
		},
	})
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
