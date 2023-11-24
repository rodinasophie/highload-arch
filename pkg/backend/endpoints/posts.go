package endpoints

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/storage"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
)

type PostCreateBody struct {
	Text string `json:"text"`
}

type PostGetBody struct {
	Id       string `json:"id"`
	AuthorId string `json:"author_user_id"`
	Text     string `json:"text"`
}

type PostUpdateBody struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

func PostCreatePost(w http.ResponseWriter, r *http.Request) {
	requestID, _ := GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	decoder := json.NewDecoder(r.Body)
	var pb PostCreateBody
	err := decoder.Decode(&pb)
	if err != nil {
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	userID, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	err = storage.CreatePost(context.Background(), userID, pb.Text)
	if err != nil {
		log.Println(err)
		GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func PostDeletePut(w http.ResponseWriter, r *http.Request) {
	requestID, _ := GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	id, ok := vars["id"]

	if !ok {
		log.Println("id is missing in parameters")
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	_, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	err = storage.DeletePost(context.Background(), id)
	if err != nil {
		log.Println(err)
		GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func PostGetGet(w http.ResponseWriter, r *http.Request) {
	requestID, _ := GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	id, ok := vars["id"]

	if !ok {
		log.Println("id is missing in parameters")
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	_, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	post, err := storage.GetPost(context.Background(), id)
	if err != nil {
		log.Println(err)
		if err == storage.ErrPostNotFound {
			GenerateError(w, http.StatusNotFound, requestID, "10m")
		} else {
			GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		}
		return
	}
	w.WriteHeader(http.StatusOK)

	resp := PostGetBody{Id: post.ID, AuthorId: post.AuthorUserID, Text: post.Text}
	json.NewEncoder(w).Encode(resp)

}

func PostUpdatePut(w http.ResponseWriter, r *http.Request) {
	requestID, _ := GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	decoder := json.NewDecoder(r.Body)
	var pb PostUpdateBody
	err := decoder.Decode(&pb)
	if err != nil {
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	_, err = CheckAuthorization(context.Background(), r)
	if err != nil {
		GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	err = storage.UpdatePost(context.Background(), pb.Id, pb.Text)
	if err != nil {
		log.Println(err)
		GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func PostFeedGet(w http.ResponseWriter, r *http.Request) {
	requestID, _ := GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	parsedQuery, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	var offset, limit int

	if len(parsedQuery) == 0 {
		offset = 0
		limit = 10
	}
	var err1, err2 error
	for key, values := range parsedQuery {
		if key == "offset" {
			offset, err1 = strconv.Atoi(values[0])
		}
		if key == "limit" {
			limit, err2 = strconv.Atoi(values[0])
		}
	}
	if err1 != nil || err2 != nil {
		log.Println(err1, err2)
		GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	userID, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	posts, err := storage.FeedPosts(context.Background(), userID, offset, limit)
	if err != nil {
		log.Println(err)
		GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)

	/* FIXME: empty array is considered as null */
	var resp []*PostGetBody
	for _, post := range posts {
		resp = append(resp, &PostGetBody{Id: post.ID, AuthorId: post.AuthorUserID, Text: post.Text})
	}
	json.NewEncoder(w).Encode(resp)

}
