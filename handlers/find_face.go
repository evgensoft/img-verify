package handlers

import (
	"fmt"
	"image"

	pigo "github.com/esimov/pigo/core"
)

var pigo1 *pigo.Pigo

func CascadeInit(cascadeFile []byte) error {
	var err error

	pigo1, err = pigo1.Unpack(cascadeFile)
	if err != nil {
		return fmt.Errorf("failed pigo unpack cascadeFile: %w", err)
	}

	return nil
}

func findFace(img image.Image) bool {
	src := pigo.ImgToNRGBA(img)

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	cParams := pigo.CascadeParams{
		MinSize:     cols / 10,
		MaxSize:     cols / 10 * 8,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}

	var qThresh float32 = 5.0

	findFace := false

	dets := pigo1.RunCascade(cParams, 1)
	for _, face := range dets {
		if face.Q > qThresh {
			return true
		}
	}
	// cascade rotation angle. 0.0 is 0 radians and 1.0 is 2*pi radians
	for angle := 0.0; angle < 0.8; angle += 0.1 {
		dets = pigo1.RunCascade(cParams, angle)
		// log.Printf("dets1 = %+#v\n", dets)

		// Calculate the intersection over union (IoU) of two clusters.
		dets = pigo1.ClusterDetections(dets, 0.02)
		// log.Printf("dets2 = %+#v\n", dets)

		for _, face := range dets {
			if face.Q > qThresh {
				return true
			}
		}
	}

	return findFace
}
