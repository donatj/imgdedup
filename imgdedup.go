package main

import (
	"flag"
	"fmt"
	"image"
	"math"
	// "image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	// "strings"
	// "strconv"
)

var subdivisions *int

func pictable(dx int, dy int) [][][]uint64 {
	pic := make([][][]uint64, dx) /* type declaration */
	for i := range pic {
		pic[i] = make([][]uint64, dy) /* again the type? */
		for j := range pic[i] {
			pic[i][j] = []uint64{0, 0, 0}
		}
	}
	return pic
}

func init() {
	subdivisions = flag.Int("subdivisions", 10, "Number of times per axis to slice image")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: imgavg [dir]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
}

func main() {

	fmt.Println(flag.NArg())

	// var imgdata [][][][]uint64
	imgdata := make(map[string][][][]uint64)

	for _, imgpath := range flag.Args() {

		// imgpath, _ = filepath.Abs(imgpath)

		if filepath.Ext(imgpath) == ".png" {
			// fmt.Println(filepath.Ext(imgpath))
			fmt.Println(imgpath)

			file, err := os.Open(imgpath)
			if err != nil {
				log.Fatal(err)
			}

			m, _, err := image.Decode(file)
			if err != nil {
				log.Fatal(err)
			}
			bounds := m.Bounds()

			avgdata := pictable(*subdivisions, *subdivisions)

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

			for rX := 0; rX < *subdivisions; rX++ {
				for rY := 0; rY < *subdivisions; rY++ {
					avgdata[rX][rY][0] = avgdata[rX][rY][0] / divisor
					avgdata[rX][rY][1] = avgdata[rX][rY][1] / divisor
					avgdata[rX][rY][2] = avgdata[rX][rY][2] / divisor
				}
			}

			// fmt.Printf("%v", avgdata)
			//
			// imgdata = append( imgdata, avgdata )

			imgdata[imgpath] = avgdata

		} else {
			fmt.Println(filepath.Ext(imgpath), "Not Supported")
		}
	}

	fmt.Printf("%v", imgdata)

}
