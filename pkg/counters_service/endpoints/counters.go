package endpoints

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/common"
	"highload-arch/pkg/counters_service/storage"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// GET /counters/{user_id}/unreadMessages
// token - user token
// from user_id to token's user
func CountersGetUnreadMessages(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	to, ok := vars["user_id"]
	if !ok {
		log.Println("user_id is missing in parameters")
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	userID := common.CheckAuth(r)
	if userID == "" {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	count, err := storage.GetMessageCount(context.Background(), userID, to)
	if err != nil {
		log.Println(err)
		common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(count)

}
