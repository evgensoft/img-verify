package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"img-verify/logger"

	"github.com/corona10/goimagehash"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

const (
	MaxPixels         = 14000
	NoteImageNotFound = "NOT_FOUND"
	NoteFaceNotFound  = "NO_FACE"
	GetImageTimeout   = 60 * time.Second
	apiRequestTimeout = 10 * time.Second
	MaxImageBytes     = 20 * 1024 * 1024
)

var (
	log         = logger.GetLogger()
	errImgur429 = errors.New("imgur StatusTooManyRequests")
)

type ResponseCloud struct {
	Results []struct {
		Results []struct {
			Classification struct {
				Properties []struct {
					Name        string  `json:"name,omitempty"`
					Probability float64 `json:"probability,omitempty"`
				} `json:"properties,omitempty"`
			} `json:"classification,omitempty"`
			FaceDetection struct {
				Faces []struct {
					BoundingBox struct {
						Vertices []struct {
							X string `json:"x,omitempty"`
							Y string `json:"y,omitempty"`
						} `json:"vertices,omitempty"`
					} `json:"boundingBox,omitempty"`
				} `json:"faces,omitempty"`
			} `json:"faceDetection,omitempty"`
		} `json:"results,omitempty"`
	} `json:"results,omitempty"`
}

func ImageInfo(msg *Message, onlyHash bool) {
	payload, err := ApiRequest("GET", msg.URL, nil)
	if err != nil {
		if err == errImgur429 { // imgur StatusTooManyRequests in YandexCloud
			return
		}

		msg.Error = err.Error()
		msg.Note = NoteImageNotFound

		return
	}

	log.Debug().Msgf("load image (%v Kb)", len(*payload)/1024)

	err = checkImageFormat(payload)
	if err != nil {
		msg.Error = err.Error()
		msg.Note = NoteImageNotFound

		return
	}

	img1, _, err := image.Decode(bytes.NewReader(*payload))
	if err != nil {
		msg.Error = err.Error()
		msg.Note = NoteImageNotFound

		return
	}

	calculateHash(msg, img1)

	if onlyHash {
		return
	}

	// Определяем работу в YANDEX_CLOUD
	yaCloud, exists := os.LookupEnv("YANDEX_CLOUD_SERVERLESS_FUNCTION")
	if !exists { // если не YANDEX_CLOUD - определяем лица локальной библиотекой
		if !findFace(img1) {
			msg.Note = NoteFaceNotFound
		}
	}

	if yaCloud == "true" {
		err := ImageYandexModeration(payload)
		if err != nil {
			msg.Error = err.Error()
			if err.Error() == NoteFaceNotFound {
				msg.Note = NoteFaceNotFound
			} else {
				msg.Note = NoteImageNotFound
			}
		}
	}
}

func calculateHash(msg *Message, img1 image.Image) {
	hash1, err := goimagehash.DifferenceHash(img1)
	if err != nil {
		msg.Error = err.Error()

		return
	}

	msg.Hash = hash1.ToString()

	log.Debug().Msgf("msg.Hash = %v", msg.Hash)

	// hash default image for "Picture not found" = "d:40e0c6a6f4008080"
	if hash1.ToString() == "d:40e0c6a6f4008080" {
		msg.Note = NoteImageNotFound

		return
	}
}

func ImageYandexModeration(payload *[]byte) error {
	if len(*payload) >= 1000000 {
		// TODO: decrease image size for check in Cloud
		return nil
	}

	reqStr := `{"folderId": "b1gdc4jel0cegsk6h65s", "analyze_specs": [{"content": "##content##","features": [{"type": "CLASSIFICATION","classificationConfig":{"model": "quality"}},{"type": "CLASSIFICATION", "classificationConfig": {"model": "moderation"}},{"type": "FACE_DETECTION"}]}]}`
	req := []byte(strings.ReplaceAll(reqStr, "##content##", base64.StdEncoding.EncodeToString(*payload)))

	body, err := ApiRequest("POST", "https://vision.api.cloud.yandex.net/vision/v1/batchAnalyze", req)
	if err != nil {
		return err
	}

	var resp ResponseCloud

	var countImg int

	log.Debug().Msgf("Response from Cloud - %s", string(*body))

	err = json.Unmarshal(*body, &resp)
	if err != nil {
		return err
	}

	// Check quality and moderation
	for _, v := range resp.Results[0].Results {
		for _, n := range v.Classification.Properties {
			if n.Name == "low" && n.Probability > 0.7 {
				return fmt.Errorf("Image low quality (%v)", n.Probability)
			}

			if n.Name == "text" && n.Probability > 0.7 {
				return fmt.Errorf("Image has text (%v)", n.Probability)
			}

			if n.Name == "watermarks" && n.Probability > 0.7 {
				return fmt.Errorf("Image has watermarks (%v)", n.Probability)
			}
		}

		if len(v.FaceDetection.Faces) > 0 {
			countImg = len(v.FaceDetection.Faces)
		}
	}

	// Check faces
	if countImg == 0 {
		return fmt.Errorf("%v", NoteFaceNotFound)
	}

	return nil
}

func checkImageFormat(img *[]byte) error {
	config, format, err := image.DecodeConfig(bytes.NewReader(*img))
	if err != nil {
		return err
	}

	log.Debug().Msgf("Format image: %v", format)

	if format == "gif" {
		err := checkGifImage(img)
		if err != nil {
			return err
		}
	}

	if config.Width+config.Height > MaxPixels {
		return fmt.Errorf("Image is too big in pixels - %v", config.Width+config.Height)
	}

	return nil
}

func checkGifImage(img *[]byte) error {
	g, err := gif.DecodeAll(bytes.NewReader(*img))
	if err != nil {
		return err
	}

	if len(g.Image) > 0 {
		return fmt.Errorf("Animated GIF")
	}

	return nil
}

func ApiRequest(method, url string, body []byte) (*[]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiRequestTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error create NewRequest - %w", err)
	}

	// Для запросов в YANDEX_CLOUD - добавляем IAMToken
	if strings.Contains(url, "api.cloud.yandex.net") {
		creds := ycsdk.InstanceServiceAccount()

		token, err := creds.IAMToken(ctx)
		if err != nil {
			return nil, err
		}

		request.Header.Set("Authorization", "Bearer "+token.IamToken)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("status code from apiRequest: %d", response.StatusCode)
		}
		// Для запросов в imgur.com выдает code 429 - rate limits
		// TODO - добавить загрузку через прокси
		if strings.Contains(url, "imgur.com") && response.StatusCode == http.StatusTooManyRequests {
			log.Debug().Msgf("Err load image from imgur.com - %v", response.Header)

			return nil, errImgur429
		}

		resBody, _ := io.ReadAll(response.Body)

		return nil, fmt.Errorf("unexpected status code from apiRequest: %d - %v", response.StatusCode, string(resBody))
	}

	response.Body = http.MaxBytesReader(nil, response.Body, MaxImageBytes)

	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &resBody, nil
}
