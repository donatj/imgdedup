package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb"
	"io"
	"log"
	"math"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

var scratchDir string

var (
	subdivisions *int = flag.Int("subdivisions", 10, "Slices per axis")
	tolerance    *int = flag.Int("tolerance", 100, "Color delta tolerance, higher = more tolerant")
)

func init() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: imgdedup [options] [<directories>/files]")
		flag.PrintDefaults()
		os.Exit(1)
	}
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

func main() {

	imgdata := make(map[string]*imageInfo)

	fileList := getFiles(flag.Args())

	bar := pb.StartNew(len(fileList))
	bar.Output = os.Stderr

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

			cacheUnit := imgpath + "|" + string(*subdivisions) + "|" + string(fi.Size()) + string(fi.ModTime().Unix())

			io.WriteString(h, cacheUnit)
			cachename := path.Join(scratchDir, fmt.Sprintf("%x", h.Sum(nil))+".tmp")

			imginfo, err := loadCache(cachename)

			if err != nil {

				imginfo, err = scanImg(file)
				if err != nil {
					log.Print(imgpath, " - ", err)
					continue
				}

				storeCache(cachename, imginfo)
			}

			imgdata[imgpath] = imginfo

			file.Close()

		} else {
			file.Close()
		}
	}

	bar.Finish()

	fileLength := len(fileList)

	for i := 0; i < fileLength-1; i++ {
		for j := i + 1; j < fileLength-1; j++ {

			filename1 := fileList[i]
			filename2 := fileList[j]

			imgdata1, ok1 := imgdata[filename1]
			imgdata2, ok2 := imgdata[filename2]

			if ok1 && ok2 {

				avgdata1 := imgdata1.Data
				avgdata2 := imgdata2.Data

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

					base := filepath.Base(fpath)
					if string(base[0]) == "." {
						return nil
					}

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

func absdiff(a uint64, b uint64) uint64 {
	return uint64(math.Abs(float64(a) - float64(b)))
}
