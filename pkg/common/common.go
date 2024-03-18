package common

import (
	"encoding/json"
	"highload-arch/pkg/config"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Auth struct {
	UserID string `json:"user_id"`
}

func GetRequestID(r *http.Request) (string, error) {
	requestID := r.Header.Get("X-Request-ID")
	return requestID, nil
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
