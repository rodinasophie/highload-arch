package common

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type ErrorResp struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Code      int    `json:"code"`
}

var ErrRequestNotAuthorized = errors.Errorf("Request not authorized")

func NewErrorResp(requestID string, code int, message string) *ErrorResp {
	return &ErrorResp{Message: message, RequestID: requestID, Code: code}
}

func GenerateError(w http.ResponseWriter, errorStatus int, requestID string, retryAfterTimeout string) {
	if retryAfterTimeout != "" {
		w.Header().Set("Retry-After", retryAfterTimeout)
	}
	resp := NewErrorResp(requestID, errorStatus, http.StatusText(errorStatus))
	w.WriteHeader(errorStatus)
	json.NewEncoder(w).Encode(resp)
}

func GenerateErrorEcho(c echo.Context, errorStatus int, requestID string, retryAfterTimeout string) error {
	if retryAfterTimeout != "" {
		c.Request().Header.Set("Retry-After", retryAfterTimeout)
	}
	resp := NewErrorResp(requestID, errorStatus, http.StatusText(errorStatus))
	return c.JSON(errorStatus, resp)
}
