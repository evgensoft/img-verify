package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"img-verify/handlers"
)

type RequestBody struct {
	HttpMethod string `json:"httpMethod"`
	Body       []byte `json:"body"`
}

type Response struct {
	StatusCode int               `json:"statusCode"`
	Body       *handlers.Message `json:"body"`
}

// Handler - функция, для запуска в Yandex Cloud Functions
func Handler(ctx context.Context, request []byte) (*Response, error) {
	requestBody := &RequestBody{}
	// Массив байтов, содержащий тело запроса, преобразуется в соответствующий объект
	err := json.Unmarshal(request, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("an error has occurred when parsing request: %v", err)
	}

	req := &handlers.Message{}
	// Поле body запроса преобразуется в объект типа Request для получения переданного имени
	err = json.Unmarshal(requestBody.Body, &req)
	if err != nil {
		return nil, fmt.Errorf("an error has occurred when parsing body: %v", err)
	}

	if req.URL == "" {
		return nil, errors.New("URL is null")
	}
	/*
		creds := ycsdk.InstanceServiceAccount()

		token, err := creds.IAMToken(ctx)
		if err != nil {
			return nil, err
		}

		log.Infof("token received - %v", token.IamToken)
		log.Infof("request received - %+#v", req)
	*/
	handlers.ImageInfo(req, true)

	log.Info().Msgf("send - %+#v", req)
	// Тело ответа необходимо вернуть в виде структуры, которая автоматически преобразуется в JSON-документ,
	// который отобразится на экране
	return &Response{
		StatusCode: 200,
		Body:       req,
	}, nil
}

func ImageYandexModeration(msg *handlers.Message) {
	type ClassificationConfig struct {
		Model string `json:"model"`
	}

	type Features struct {
		Type                 string               `json:"type"`
		ClassificationConfig ClassificationConfig `json:"classificationConfig"`
	}

	type AnalyzeSpecs struct {
		Content  string     `json:"content"`
		Features []Features `json:"features"`
	}

	type ImageYaModeration struct {
		FolderID     string         `json:"folderId"`
		AnalyzeSpecs []AnalyzeSpecs `json:"analyze_specs"`
	}

	payload, err := handlers.ApiRequest("GET", msg.URL, nil)
	if err != nil {
		msg.Error = err.Error()
		msg.Note = handlers.NoteImageNotFound

		return
	}

	imgB64 := base64.StdEncoding.EncodeToString(*payload)
	a := AnalyzeSpecs{}
	a.Content = imgB64

	a.Features = append(a.Features, Features{Type: "Classification", ClassificationConfig: ClassificationConfig{Model: "quality"}})
	yaMod := ImageYaModeration{}
	yaMod.FolderID = "b1gdc4jel0cegsk6h65s"
	yaMod.AnalyzeSpecs = append(yaMod.AnalyzeSpecs, a)
}
