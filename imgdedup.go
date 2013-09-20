package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

var subdivisions *int
var cutoff *int

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

func absdiff(a uint64, b uint64) uint64 {
	return uint64(math.Abs(float64(a) - float64(b)))
}

func init() {
	subdivisions = flag.Int("subdivisions", 10, "Number of times per axis to slice image")
	cutoff = flag.Int("cutoff", 100, "Cutoff to declare similar")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: imgavg [dir]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func getFiles(paths []string) []string {
	var theFiles []string

	for _, imgpath := range paths {

		file, err := os.Open(imgpath)
		if err != nil {
			log.Fatal(err)
		}

		fi, err := file.Stat()
		if err != nil {
			log.Fatal(err)
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			// fmt.Println("directory")
			filepath.Walk(imgpath, func(path string, f os.FileInfo, err error) error {

				submode := f.Mode()
				if submode.IsRegular() {
					theFiles = append(theFiles, path)
				}

				return nil
			})
		case mode.IsRegular():
			// fmt.Println("file")
			theFiles = append(theFiles, imgpath)
		}

		file.Close()

	}

	return theFiles
}

func main() {
	// var imgdata [][][][]uint64
	imgdata := make(map[string][][][]uint64)

	theFiles := getFiles(flag.Args())

	for _, imgpath := range theFiles {

		file, err := os.Open(imgpath)
		if err != nil {
			log.Fatal(err)
		}
		// imgpath, _ = filepath.Abs(imgpath)

		if filepath.Ext(imgpath) == ".png" || filepath.Ext(imgpath) == ".jpg" || filepath.Ext(imgpath) == ".jpeg" {

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

			imgdata[imgpath] = avgdata

			file.Close()

		} else {
			fmt.Println(filepath.Ext(imgpath), "Not Supported")
		}
	}

	for filename1, avgdata1 := range imgdata {
		for filename2, avgdata2 := range imgdata {
			if filename1 == filename2 {
				continue
			}

			var xdiff uint64 = 0

			for rX := 0; rX < *subdivisions; rX++ {
				for rY := 0; rY < *subdivisions; rY++ {
					aa := avgdata1[rX][rY]
					bb := avgdata2[rX][rY]

					xdiff += absdiff(absdiff(absdiff(aa[0], bb[0]), absdiff(aa[1], bb[1])), absdiff(aa[2], bb[2]))
				}
			}

			if xdiff < uint64(*cutoff) {
				fmt.Println(filename1, filename2)
				fmt.Println(xdiff)
			}

		}
	}

}
