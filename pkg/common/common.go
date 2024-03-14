package common

import (
	"net/http"
)

func GetRequestID(r *http.Request) (string, error) {
	requestID := r.Header.Get("X-Request-ID")
	return requestID, nil
}
