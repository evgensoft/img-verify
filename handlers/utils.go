package handlers

import (
	"bytes"
	"context"
	"image"
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
	MaxImageBytes     = 10 * 1024 * 1024
)

var log = logger.GetLogger()

func ImageInfo(msg *Message, onlyHash bool) {
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
	if err != nil {
		msg.Error = err.Error()
		msg.Note = NoteImageNotFound

		return
	}

	img1, _, err := image.Decode(bytes.NewReader(payload))
	if err != nil {
		msg.Error = err.Error()
		msg.Note = NoteImageNotFound

		return
	}

	if img1.Bounds().Size().X+img1.Bounds().Size().Y > MaxPixels {
		msg.Note = NoteImageNotFound

		log.Error("Image is too big in pixels - ", img1.Bounds().Size().X+img1.Bounds().Size().Y)

		return
	}

	hash1, err := goimagehash.DifferenceHash(img1)
	if err != nil {
		msg.Error = err.Error()

		return
	}

	msg.Hash = hash1.ToString()
	// hash default image for "Picture not fond" = "d:40e0c6a6f4008080"
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
