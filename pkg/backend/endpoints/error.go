package endpoints

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
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
