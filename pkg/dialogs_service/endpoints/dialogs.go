package endpoints

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/common"
	"highload-arch/pkg/config"
	"highload-arch/pkg/dialogs_service/storage"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

type Auth struct {
	UserID string `json:"user_id"`
}

type DialogSendBody struct {
	Text string `json:"text"`
}
type DialogListBody struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

func CheckAuth(r *http.Request) string {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) <= 1 {
		log.Println("Unauthorized")
		return ""
	}
	reqToken = splitToken[1]

	url := url.URL{}
	url.Host = config.GetString("server.host")
	url.Scheme = "http"
	url.Path = "/api/v2/checkAuth"

	proxyReq, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Println(err)
		return ""
	}
	proxyReq.Header.Set("Authorization", "Bearer "+reqToken)
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	decoder := json.NewDecoder(resp.Body)
	var auth Auth
	err = decoder.Decode(&auth)
	if err != nil {
		return ""
	}
	return auth.UserID
}

func DialogUserIdSendMessage(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	decoder := json.NewDecoder(r.Body)
	var dialog DialogSendBody
	err := decoder.Decode(&dialog)
	if err != nil {
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	vars := mux.Vars(r)
	to, ok := vars["user_id"]
	if !ok {
		log.Println("user_id is missing in parameters")
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	userID := CheckAuth(r)
	if userID == "" {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	err = storage.SendMessage(context.Background(), userID, to, dialog.Text)
	if err != nil {
		log.Println(err)
		common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)

}

func DialogUserIdListGet(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	to, ok := vars["user_id"]

	if !ok {
		log.Println("user_id is missing in parameters")
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	userID := CheckAuth(r)
	if userID != "" {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	dialog, err := storage.DialogList(context.Background(), userID, to)
	if err != nil {
		log.Println(err)
		common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)

	var resp []*DialogListBody
	for _, message := range dialog {
		resp = append(resp, &DialogListBody{From: message.AuthorID, To: message.RecepientID, Text: message.Text})
	}
	json.NewEncoder(w).Encode(resp)
}
