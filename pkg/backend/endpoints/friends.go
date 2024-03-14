package endpoints

import (
	"context"
	"highload-arch/pkg/common"
	"highload-arch/pkg/storage"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func FriendAddPut(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	friendID, ok := vars["user_id"]
	if !ok {
		log.Println("user_id is missing in parameters")
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	userID, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}
	if friendID == userID {
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	err = storage.AddFriend(context.Background(), userID, friendID)
	if err != nil {
		log.Println(err)
		if err == common.ErrUserNotFound {
			common.GenerateError(w, http.StatusNotFound, requestID, "10m")
		} else {
			common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

func FriendDeletePut(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	friendID, ok := vars["user_id"]
	if !ok {
		log.Println("user_id is missing in parameters")
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	userID, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}
	if friendID == userID {
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	err = storage.DeleteFriend(context.Background(), userID, friendID)
	if err != nil {
		log.Println(err)
		if err == common.ErrUserNotFound {
			common.GenerateError(w, http.StatusNotFound, requestID, "10m")
		} else {
			common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}
