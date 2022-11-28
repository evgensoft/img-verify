package handlers

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"io"
	"net/http"
	"time"

	"img-verify/logger"

	"github.com/corona10/goimagehash"
)

const (
	MaxPixels         = 14000
	NoteImageNotFound = "NOT_FOUND"
	NoteFaceNotFound  = "NO_FACE"
	GetImageTimeout   = 60 * time.Second
	apiRequestTimeout = 10 * time.Second
	MaxImageBytes     = 10 * 1024 * 1024
)

var log = logger.GetLogger()

func ImageInfo(msg *Message, onlyHash bool) {
	/*
		ctx, cancel := context.WithTimeout(context.Background(), GetImageTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, msg.URL, nil)
		if err != nil {
			msg.Error = err.Error()

			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			msg.Error = err.Error()

			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			msg.Error = "Ошибка загрузки файла - HTTP code " + resp.Status
			msg.Note = NoteImageNotFound

			return
		}

		resp.Body = http.MaxBytesReader(nil, resp.Body, MaxImageBytes)

		payload, err := io.ReadAll(resp.Body)
	*/
	payload, err := ApiRequest("GET", msg.URL, nil)
	if err != nil {
		msg.Error = err.Error()
		msg.Note = NoteImageNotFound

		return
	}

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

	hash1, err := goimagehash.DifferenceHash(img1)
	if err != nil {
		msg.Error = err.Error()

		return
	}

	msg.Hash = hash1.ToString()
	// hash default image for "Picture not found" = "d:40e0c6a6f4008080"
	if hash1.ToString() == "d:40e0c6a6f4008080" {
		msg.Note = NoteImageNotFound

		return
	}

	if !onlyHash {
		if !findFace(img1) {
			msg.Note = NoteFaceNotFound
		}
	}
}

func checkImageFormat(img *[]byte) error {
	config, format, err := image.DecodeConfig(bytes.NewReader(*img))
	if err != nil {
		return err
	}

	if format == "gif" {
		g, err := gif.DecodeAll(bytes.NewReader(*img))
		if err != nil {
			return err
		}

		if len(g.Image) > 0 {
			return fmt.Errorf("Animated GIF")
		}
	}

	if config.Width+config.Height > MaxPixels {
		return fmt.Errorf("Image is too big in pixels - %v", config.Width+config.Height)
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

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from apiRequest: %d", response.StatusCode)
	}

	response.Body = http.MaxBytesReader(nil, response.Body, MaxImageBytes)

	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &resBody, nil
}
