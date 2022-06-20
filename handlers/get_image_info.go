package handlers

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
)

type Message struct {
	URL   string `json:"imageurl"`
	Hash  string `json:"hash,omitempty"`
	Note  string `json:"note,omitempty"`
	Error string `json:"error,omitempty"`
}

func GetImageInfo(resp http.ResponseWriter, req *http.Request) {
	if req.Body != nil {
		defer req.Body.Close()
	}

	var reqMessage Message

	err := json.NewDecoder(req.Body).Decode(&reqMessage)
	if err != nil {
		log.Errorf("error decode json: %v", err)
		http.Error(resp, err.Error(), http.StatusBadRequest)

		return
	}

	log.Infof("Req %s %s - %+#v", req.Method, req.RequestURI, reqMessage)

	if reqMessage.URL == "" {
		http.Error(resp, "URL is null", http.StatusBadRequest)

		return
	}

	ImageInfo(&reqMessage, false)

	body, err := json.Marshal(&reqMessage)
	if err != nil {
		log.Errorf("error json.Marshal: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)

		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Header().Set("Content-Type", "application/json")

	log.Infof("Ans %s", string(body))

	_, err = resp.Write(body)
	if err != nil {
		log.Errorf("error write answer - %v", err)
	}

	debug.FreeOSMemory()
}

func GetImageHash(resp http.ResponseWriter, req *http.Request) {
	if req.Body != nil {
		defer req.Body.Close()
	}

	var reqMessage Message

	err := json.NewDecoder(req.Body).Decode(&reqMessage)
	if err != nil {
		log.Errorf("error decode json: %v", err)
		http.Error(resp, err.Error(), http.StatusBadRequest)

		return
	}

	log.Infof("Req %s %s - %+#v", req.Method, req.RequestURI, reqMessage)

	if reqMessage.URL == "" {
		http.Error(resp, "URL is null", http.StatusBadRequest)

		return
	}

	ImageInfo(&reqMessage, true)

	body, err := json.Marshal(&reqMessage)
	if err != nil {
		log.Errorf("error json.Marshal: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)

		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Header().Set("Content-Type", "application/json")

	log.Infof("Ans %s", string(body))

	_, err = resp.Write(body)
	if err != nil {
		log.Errorf("error write answer - %v", err)
	}

	debug.FreeOSMemory()
}
