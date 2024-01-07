package endpoints

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/common"
	"highload-arch/pkg/storage"
	"net/http"
	"strings"
)

type Auth struct {
	UserID string `json:"user_id"`
}

const DateFormat = "2006-01-02"

func CheckAuthorization(ctx context.Context, r *http.Request) (string, error) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) <= 1 {
		return "", common.ErrRequestNotAuthorized
	}
	reqToken = splitToken[1]

	userID, err := storage.ValidateLoginToken(ctx, reqToken)
	if err != nil {
		return "", common.ErrRequestNotAuthorized
	}
	return userID, nil
}

func CheckAuthGet(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	userID, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)
	resp := &Auth{UserID: userID}
	json.NewEncoder(w).Encode(resp)
}
