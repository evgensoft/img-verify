package handlers

import (
	"context"
	"image"
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
	MaxImageBytes     = 50 * 1024 * 1024
)

var log = logger.GetLogger()

func ImageInfo(msg *Message) {
	ctx, cancel := context.WithTimeout(context.Background(), GetImageTimeout)
	defer cancel()

	res, err := http.NewRequestWithContext(ctx, http.MethodGet, msg.URL, nil)
	if err != nil {
		msg.Error = err.Error()

		return
	}
	defer res.Body.Close()

	if res.Response.StatusCode != http.StatusOK {
		msg.Error = "Ошибка загрузки файла - HTTP code " + res.Response.Status
		msg.Note = NoteImageNotFound

		return
	}

	img1, _, err := image.Decode(http.MaxBytesReader(nil, res.Body, MaxImageBytes))
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

	if !findFace(img1) {
		msg.Note = NoteFaceNotFound
	}
}
