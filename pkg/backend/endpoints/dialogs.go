package endpoints

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/storage"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type DialogSendBody struct {
	Text string `json:"text"`
}
type DialogListBody struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

func DialogUserIdSendPost(w http.ResponseWriter, r *http.Request) {
	requestID, _ := GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	decoder := json.NewDecoder(r.Body)
	var dialog DialogSendBody
	err := decoder.Decode(&dialog)
	if err != nil {
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	vars := mux.Vars(r)
	to, ok := vars["user_id"]
	if !ok {
		log.Println("user_id is missing in parameters")
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	userID, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}
	_, err = storage.GetUser(context.Background(), to)
	if err != nil {
		log.Println(err)
		if err == storage.ErrUserNotFound {
			GenerateError(w, http.StatusNotFound, requestID, "10m")
		} else {
			GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		}
		return
	}

	err = storage.SendMessage(context.Background(), userID, to, dialog.Text)
	if err != nil {
		log.Println(err)
		GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)

}

func DialogUserIdListGet(w http.ResponseWriter, r *http.Request) {
	requestID, _ := GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	to, ok := vars["user_id"]

	if !ok {
		log.Println("user_id is missing in parameters")
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	userID, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	dialog, err := storage.DialogList(context.Background(), userID, to)
	if err != nil {
		log.Println(err)
		GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}

	var resp []*DialogListBody
	for _, message := range dialog {
		resp = append(resp, &DialogListBody{From: message.AuthorID, To: message.RecepientID, Text: message.Text})
	}
	json.NewEncoder(w).Encode(resp)

	w.WriteHeader(http.StatusOK)
}
