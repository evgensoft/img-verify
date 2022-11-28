package main

import (
	"context"
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

	handlers.ImageInfo(req, false)

	// log.Info().Msgf("send - %+#v", req)
	// Тело ответа необходимо вернуть в виде структуры, которая автоматически преобразуется в JSON-документ,
	// который отобразится на экране
	return &Response{
		StatusCode: 200,
		Body:       req,
	}, nil
}
