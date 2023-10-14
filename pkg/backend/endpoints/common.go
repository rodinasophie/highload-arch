package endpoints

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/storage"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const DateFormat = "2006-01-02"

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

func GetRequestIDEcho(c echo.Context) (string, error) {
	requestID := c.Request().Header.Get("X-Request-ID")
	return requestID, nil
}

type RequestParams map[string]interface{}

func ParseRequest(c echo.Context) (params RequestParams, err error) {
	req := c.Request()
	if req != nil {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			err = json.Unmarshal(data, &params)
			if err != nil {
				return nil, err
			}
		}
	}

	return params, nil
}
