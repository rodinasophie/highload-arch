package endpoints

import (
	"highload-arch/pkg/config"
	"log"
	"net/http"
)

type DialogSendBody struct {
	Text string `json:"text"`
}
type DialogListBody struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

func DialogUserIdSendMessage(w http.ResponseWriter, req *http.Request) {
	url := req.URL
	url.Host = config.GetString("dialogs.host")
	url.Scheme = "http"

	proxyReq, err := http.NewRequest(req.Method, url.String(), req.Body)
	if err != nil {
		log.Println(err)
		return
	}

	proxyReq.Header.Set("Host", req.Host)
	proxyReq.Header.Set("X-Forwarded-For", req.RemoteAddr)
	for header, values := range req.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	client := &http.Client{}
	_, err = client.Do(proxyReq)
	if err != nil {
		log.Println(err)
	}
}

func DialogUserIdListGet(w http.ResponseWriter, req *http.Request) {
	url := req.URL
	url.Host = config.GetString("dialogs.host")
	url.Scheme = "http"
	proxyReq, err := http.NewRequest(req.Method, url.String(), req.Body)
	if err != nil {
		log.Println(err)
		return
	}

	proxyReq.Header.Set("Host", req.Host)
	proxyReq.Header.Set("X-Forwarded-For", req.RemoteAddr)

	for header, values := range req.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	client := &http.Client{}
	_, err = client.Do(proxyReq)
	if err != nil {
		log.Println(err)
	}
}
