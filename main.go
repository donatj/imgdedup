package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/dustin/go-humanize"
	homedir "github.com/mitchellh/go-homedir"
)

var scratchDir string

var (
	subdivisions = flag.Int("subdivisions", 10, "Slices per axis")
	tolerance    = flag.Int("tolerance", 100, "Color delta tolerance, higher = more tolerant")
	difftool     = flag.String("diff", "", "Command to pass dupe images to eg: cmd $left $right")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] [<directories>/files]:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}
}

func init() {
	h, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	scratchDir = filepath.Join(h, ".imgdedup")

	if _, err := os.Stat(scratchDir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(scratchDir, 0700)
		} else {
			log.Fatal(err)
		}
	}
}

func fileData(imgpath string) (*imageInfo, error) {
	fExt := strings.ToLower(filepath.Ext(imgpath))
	if fExt == ".png" || fExt == ".jpg" || fExt == ".jpeg" || fExt == ".gif" || fExt == ".bmp" || fExt == ".webp" {
		file, err := os.Open(imgpath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		fi, err := file.Stat()
		if err != nil {
			return nil, err
		}

		h := md5.New()

		cacheUnit := getCacheName(imgpath, fi)

		io.WriteString(h, cacheUnit)
		cachename := filepath.Join(scratchDir, fmt.Sprintf("%x", h.Sum(nil))+".tmp")

		imginfo, err := loadCache(cachename)

		if err != nil {
			img, _, err := image.Decode(file)
			if err != nil {
				return nil, err
			}

			imginfo, err = scanImg(img)
			if err != nil {
				return nil, err
			}

			fi, err := file.Stat()
			if err != nil {
				return nil, err
			}

			imginfo.Filesize = uint64(fi.Size())

			err = storeCache(cachename, imginfo)
			if err != nil {
				return nil, err
			}
		}

		return imginfo, nil
	}

	return nil, fmt.Errorf("Ext %s unhandled", fExt)
}

func main() {
	fileList, err := getFiles(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	bar := pb.StartNew(len(fileList))
	bar.Output = os.Stderr

	imgdata := make(map[string]*imageInfo)
	for _, imgpath := range fileList {
		bar.Increment()
		imginfo, err := fileData(imgpath)
		if err != nil {
			//			log.Println(err)
			continue
		}

		imgdata[imgpath] = imginfo
	}

	bar.Finish()

	fileLength := len(fileList)

	for i := 0; i < fileLength; i++ {
		for j := i + 1; j < fileLength; j++ {

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

				xdiff := getDiff(avgdata1, avgdata2)

				if xdiff < uint64(*tolerance) {

					fmt.Println(filename1)
					fmt.Printf("    %d x %d\n    %s\n", imgdata1.Bounds.Dx(), imgdata1.Bounds.Dy(), humanize.Bytes(imgdata1.Filesize))

					fmt.Println(filename2)
					fmt.Printf("    %d x %d\n    %s\n", imgdata2.Bounds.Dx(), imgdata2.Bounds.Dy(), humanize.Bytes(imgdata2.Filesize))

					fmt.Println("")
					fmt.Println("Diff: ", xdiff)

					if xdiff > 0 && imgdata1.Filesize != imgdata2.Filesize {
						if *difftool != "" {
							log.Println("Launching difftool")
							cmd := exec.Command(*difftool, filename1, filename2)
							cmd.Run()
							time.Sleep(500 * time.Millisecond)

							// lots of difftools return a variety of exit codes so I can't really test for errors
							//if e, ok := err.(*exec.ExitError); ok {
							//	log.Fatal(e)
							//}
						}
					}

					fmt.Println("- - - - - - - - - -")
				}

			}

		}
	}

}

func getDiff(avgdata1 pictable, avgdata2 pictable) uint64 {
	var xdiff uint64
	for rX := 0; rX < *subdivisions; rX++ {
		for rY := 0; rY < *subdivisions; rY++ {
			aa := avgdata1[rX][rY]
			bb := avgdata2[rX][rY]

			xdiff += absdiff(absdiff(absdiff(aa[0], bb[0]), absdiff(aa[1], bb[1])), absdiff(aa[2], bb[2]))
		}
	}
	return xdiff
}

func getFiles(paths []string) ([]string, error) {
	var fileList []string

	for _, imgpath := range paths {

		file, err := os.Open(imgpath)
		if err != nil {
			return fileList, err
		}

		fi, err := file.Stat()
		if err != nil {
			return fileList, err
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			// Walk is recursive
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
			fpath, _ := filepath.Abs(imgpath)
			fileList = append(fileList, fpath)
		}

		file.Close()

	}

	return fileList, nil
}

func absdiff(a uint64, b uint64) uint64 {
	return uint64(math.Abs(float64(a) - float64(b)))
}
