package endpoints

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/common"
	"highload-arch/pkg/storage"
	"log"
	"net/http"
)

type LoginBody struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

type LoginResp struct {
	Token string `json:"token"`
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	decoder := json.NewDecoder(r.Body)
	var rb LoginBody
	err := decoder.Decode(&rb)
	if err != nil {
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	/* Get user and return error if the user doesn't exist */
	_, err = storage.GetUser(context.Background(), rb.ID)
	if err != nil {
		log.Println(err)
		if err == common.ErrUserNotFound {
			common.GenerateError(w, http.StatusNotFound, requestID, "10m")
		} else {
			common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		}
		return
	}

	/* Login the  user if the user exists */
	login, err := storage.LoginUser(context.Background(), &storage.Login{rb.ID, rb.Password})
	if err == common.ErrPasswordInvalid {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}
	if err != nil {
		log.Println(err)
		common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)
	resp := &LoginResp{Token: login.Token}
	json.NewEncoder(w).Encode(resp)
}
