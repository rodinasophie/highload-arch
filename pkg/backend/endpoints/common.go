package endpoints

import (
	"context"
	"highload-arch/pkg/storage"
	"net/http"
	"strings"
)

func CheckAuthorization(ctx context.Context, r *http.Request, userID string) error {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ") // FIXME: to Token
	if len(splitToken) <= 1 {
		return ErrRequestNotAuthorized
	}
	reqToken = splitToken[1]

	if storage.CheckLoginToken(ctx, reqToken, userID) != nil {
		return ErrRequestNotAuthorized
	}
	return nil
}

func GetRequestID(r *http.Request) (string, error) {
	requestID := r.Header.Get("X-Request-ID")
	return requestID, nil
}
