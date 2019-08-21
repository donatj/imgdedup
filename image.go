package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"strings"

	// Format self registers

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type pictable [][][3]uint64

func newPictable(dx int, dy int) pictable {
	pic := make([][][3]uint64, dx) /* type declaration */
	for i := range pic {
		pic[i] = make([][3]uint64, dy) /* again the type? */
		for j := range pic[i] {
			pic[i][j] = [3]uint64{0, 0, 0}
		}
	}
	return pic
}

type imageInfo struct {
	Data     pictable
	Format   string
	Bounds   image.Rectangle
	Filesize uint64
}

func newImageInfo(imgpath string) (*imageInfo, error) {
	fExt := strings.ToLower(filepath.Ext(imgpath))
	if !(fExt == ".png" || fExt == ".jpg" || fExt == ".jpeg" || fExt == ".gif" || fExt == ".bmp" || fExt == ".webp" || fExt == ".tiff") {
		return nil, fmt.Errorf("Ext %s unhandled", fExt)
	}

	file, err := os.Open(imgpath)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	img, ifmt, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	pict, err := newPictableFromImage(img)
	if err != nil {
		return nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	imginfo := &imageInfo{
		Data:     pict,
		Format:   ifmt,
		Bounds:   img.Bounds(),
		Filesize: uint64(fi.Size()),
	}

	return imginfo, nil

}

func newPictableFromImage(m image.Image) (pictable, error) {
	bounds := m.Bounds()

	avgdata := newPictable(*subdivisions, *subdivisions)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rX := int64(math.Floor((float64(x) / float64(bounds.Max.X)) * float64(*subdivisions)))
			rY := int64(math.Floor((float64(y) / float64(bounds.Max.Y)) * float64(*subdivisions)))

			r, g, b, _ := m.At(x, y).RGBA()
			avgdata[rX][rY][0] += uint64((float32(r) / 65535) * 255)
			avgdata[rX][rY][1] += uint64((float32(g) / 65535) * 255)
			avgdata[rX][rY][2] += uint64((float32(b) / 65535) * 255)
		}
	}

	divisor := uint64((bounds.Max.X / *subdivisions) * (bounds.Max.Y / *subdivisions))
	if divisor == 0 {
		return nil, fmt.Errorf("Image dimensions %d x %d invalid", bounds.Max.X, bounds.Max.Y)
	}

	for rX := 0; rX < *subdivisions; rX++ {
		for rY := 0; rY < *subdivisions; rY++ {
			avgdata[rX][rY][0] = avgdata[rX][rY][0] / divisor
			avgdata[rX][rY][1] = avgdata[rX][rY][1] / divisor
			avgdata[rX][rY][2] = avgdata[rX][rY][2] / divisor
		}
	}

	return avgdata, nil
}
