package endpoints

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/common"
	"highload-arch/pkg/storage"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	websocket "github.com/gorilla/websocket"
)

const WEBSOCKET_ENABLED = true

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
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	decoder := json.NewDecoder(r.Body)
	var pb PostCreateBody
	err := decoder.Decode(&pb)
	if err != nil {
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	userID, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	err = storage.CreatePost(context.Background(), userID, pb.Text)
	if err != nil {
		log.Println(err)
		common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func PostDeletePut(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	id, ok := vars["id"]

	if !ok {
		log.Println("id is missing in parameters")
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	_, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	err = storage.DeletePost(context.Background(), id)
	if err != nil {
		log.Println(err)
		common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func PostGetGet(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	id, ok := vars["id"]

	if !ok {
		log.Println("id is missing in parameters")
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	_, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	post, err := storage.GetPost(context.Background(), id)
	if err != nil {
		log.Println(err)
		if err == common.ErrPostNotFound {
			common.GenerateError(w, http.StatusNotFound, requestID, "10m")
		} else {
			common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		}
		return
	}
	w.WriteHeader(http.StatusOK)

	resp := PostGetBody{Id: post.ID, AuthorId: post.AuthorUserID, Text: post.Text}
	json.NewEncoder(w).Encode(resp)

}

func PostUpdatePut(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	decoder := json.NewDecoder(r.Body)
	var pb PostUpdateBody
	err := decoder.Decode(&pb)
	if err != nil {
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}

	_, err = CheckAuthorization(context.Background(), r)
	if err != nil {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	err = storage.UpdatePost(context.Background(), pb.Id, pb.Text)
	if err != nil {
		log.Println(err)
		common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func sendPostViaWebsocket(conn *websocket.Conn, post []byte) {
	log.Println("Sending message to websocket")
	if err := conn.WriteMessage(1, post); err != nil {
		log.Println(err)
		return
	}
}

func PostFeedGetWebsocket(w http.ResponseWriter, r *http.Request) {
	userID, err := CheckAuthorization(context.Background(), r)

	if err != nil {
		log.Println("Authorization failed for websocket")
		return
	}
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
		return
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection

	err = ReadPostCreatedMessageFromQueue(context.Background(), userID, sendPostViaWebsocket, ws)
	if err != nil {
		log.Println("Cannot read messages from queue on the client side")
		return
	}
}

func PostFeedGet(w http.ResponseWriter, r *http.Request) {
	requestID, _ := common.GetRequestID(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	parsedQuery, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
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
		common.GenerateError(w, http.StatusBadRequest, requestID, "10m")
		return
	}
	userID, err := CheckAuthorization(context.Background(), r)
	if err != nil {
		common.GenerateError(w, http.StatusUnauthorized, requestID, "10m")
		return
	}

	posts, err := storage.FeedPosts(context.Background(), userID, offset, limit)
	if err != nil {
		log.Println(err)
		common.GenerateError(w, http.StatusInternalServerError, requestID, "10m")
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
