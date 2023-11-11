package endpoints

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/storage"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

type UserRegisterBody struct {
	FirstName  string `json:"first_name,omitempty"`
	SecondName string `json:"second_name,omitempty"`
	Birthdate  string `json:"birthdate,omitempty"`
	Biography  string `json:"biography,omitempty"`
	City       string `json:"city,omitempty"`
	Password   string `json:"password,omitempty"`
}

type UserRegisterResponse struct {
	UserID string `json:"user_id"`
}

type UserGetResponse struct {
	FirstName  string `json:"first_name,omitempty"`
	SecondName string `json:"second_name,omitempty"`
	Birthdate  string `json:"birthdate,omitempty"`
	Biography  string `json:"biography,omitempty"`
	City       string `json:"city,omitempty"`
}

type UserGetResponseID struct {
	ID         string `json:id`
	FirstName  string `json:"first_name,omitempty"`
	SecondName string `json:"second_name,omitempty"`
	Birthdate  string `json:"birthdate,omitempty"`
	Biography  string `json:"biography,omitempty"`
	City       string `json:"city,omitempty"`
}

func UserGetIdGet(w http.ResponseWriter, r *http.Request) {
	requestID, _ := GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	userID, ok := vars["id"]
	if !ok {
		log.Println("id is missing in parameters")
	}
	var err error
	if err != nil {
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	_, err = CheckAuthorization(context.Background(), r)
	if err != nil {
		GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}
	var user *storage.User
	user, err = storage.GetUser(context.Background(), userID)
	if err != nil {
		log.Println(err)
		if err == storage.ErrUserNotFound {
			GenerateError(w, http.StatusNotFound, requestID, "10m")
		} else {
			GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	resp := &UserGetResponse{user.FirstName, user.SecondName, user.Birthdate.Format("2006-01-02"), user.Biography, user.City}
	json.NewEncoder(w).Encode(resp)
}

func UserRegisterPost(w http.ResponseWriter, r *http.Request) {
	requestID, _ := GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	decoder := json.NewDecoder(r.Body)
	var rb UserRegisterBody
	err := decoder.Decode(&rb)
	if err != nil {
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	birthdate, err := time.Parse(time.DateOnly, rb.Birthdate)
	if err != nil {
		log.Println(err)
		GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	id, err := storage.AddUser(context.Background(), &storage.User{ID: "", FirstName: rb.FirstName, SecondName: rb.SecondName, Birthdate: birthdate, Biography: rb.Biography, City: rb.City}, rb.Password)
	if err != nil {
		log.Println(err)
		GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)
	resp := &UserRegisterResponse{UserID: id}
	json.NewEncoder(w).Encode(resp)
}

func UserSearchGet(w http.ResponseWriter, r *http.Request) {
	requestID, _ := GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	parsedQuery, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	var firstName, secondName string
	for key, values := range parsedQuery {
		if key == "first_name" {
			firstName = values[0]
		}
		if key == "second_name" {
			secondName = values[0]
		}
	}
	users, err := storage.SearchUsers(context.Background(), firstName, secondName)
	if err != nil {
		log.Println(err)
		if err == storage.ErrUserNotFound {
			GenerateError(w, http.StatusNotFound, requestID, "10m")
		} else {
			GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	var resp []*UserGetResponseID
	for _, user := range users {
		resp = append(resp, &UserGetResponseID{user.ID, user.FirstName, user.SecondName, user.Birthdate.Format(DateFormat), user.Biography, user.City})
	}
	json.NewEncoder(w).Encode(resp)
}
