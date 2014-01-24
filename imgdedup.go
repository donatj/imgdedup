package main

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

var subdivisions *int
var tolerance *int

var scratchDir string

type pictable [][][]uint64

func MkPictable(dx int, dy int) pictable {
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
	subdivisions = flag.Int("subdivisions", 10, "Slices per axis")
	tolerance = flag.Int("tolerance", 100, "Color delta tolerance, higher = more tolerant")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: imgdedup [options] [<directories>/files]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
}

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	scratchDir = path.Join(usr.HomeDir, ".imgdedup")

	if _, err := os.Stat(scratchDir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(scratchDir, 0700)
		} else {
			log.Fatal(err)
		}
	}
}

func scanImg(file *os.File) (pictable, error) {
	m, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := m.Bounds()

	avgdata := MkPictable(*subdivisions, *subdivisions)

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

	return avgdata, nil
}

func loadCache(cachename string) (pictable, error) {

	file, err := os.Open(cachename)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(file)

	var avgdata pictable

	dec := json.NewDecoder(r)

	err = dec.Decode(&avgdata)
	if err != nil {
		return nil, err
	}

	return avgdata, nil
}

func storeCache(cachename string, avgdata *pictable) {
	fo, err := os.Create(cachename)
	if err != nil {
		panic(err)
	}

	enc := json.NewEncoder(fo)
	enc.Encode(avgdata)
	fo.Close()
}

func main() {
	// var imgdata [][][][]uint64
	imgdata := make(map[string][][][]uint64)

	fileList := getFiles(flag.Args())

	bar := pb.StartNew(len(fileList))

	for _, imgpath := range fileList {

		bar.Increment()

		file, err := os.Open(imgpath)
		if err != nil {
			log.Fatal(err)
		}

		fExt := strings.ToLower(filepath.Ext(imgpath))
		if fExt == ".png" || fExt == ".jpg" || fExt == ".jpeg" || fExt == ".gif" {

			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			h := md5.New()
			io.WriteString(h, imgpath+"|"+string(*subdivisions)+"|"+string(fi.Size())+string(fi.ModTime().UnixNano()))
			cachename := path.Join(scratchDir, fmt.Sprintf("%x", h.Sum(nil))+".tmp")

			var avgdata pictable

			avgdata, err = loadCache(cachename)
			if err != nil {

				avgdata, err = scanImg(file)
				if err != nil {
					log.Print(imgpath, err)
					continue
				}

				storeCache(cachename, &avgdata)
			}

			imgdata[imgpath] = avgdata

			file.Close()

		} else {
			file.Close()
		}
	}

	fileLength := len(fileList)

	for i := 0; i < fileLength-1; i++ {
		for j := i + 1; j < fileLength-1; j++ {

			filename1 := fileList[i]
			filename2 := fileList[j]

			avgdata1, ok1 := imgdata[filename1]
			avgdata2, ok2 := imgdata[filename2]

			if ok1 && ok2 {

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

				if xdiff < uint64(*tolerance) {
					fmt.Println(filename1, filename2)
					fmt.Println(xdiff)
				}

			}

		}
	}

}

func getFiles(paths []string) []string {
	var fileList []string

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
					fpath, _ := filepath.Abs(path)
					fileList = append(fileList, fpath)
				}

				return nil
			})
		case mode.IsRegular():
			// fmt.Println("file")
			fpath, _ := filepath.Abs(imgpath)
			fileList = append(fileList, fpath)
		}

		file.Close()

	}

	return fileList
}
