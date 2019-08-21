package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/dustin/go-humanize"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/prologic/bitcask"
)

var (
	subdivisions = flag.Int("subdivisions", 10, "Slices per axis")
	tolerance    = flag.Int("tolerance", 100, "Color delta tolerance, higher = more tolerant")
	difftool     = flag.String("diff", "", "Command to pass dupe images to eg: cmd $left $right")
)

var (
	cacheDir *string
	cacheDb  *bitcask.Bitcask
)

func init() {
	h, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	cacheDir = flag.String("cache-dir", filepath.Join(h, ".imgdedup/cacheDb"), "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] [<directories>/files]:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	cacheDb, err = bitcask.Open(*cacheDir)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer cacheDb.Close()
	cacheDb.Put("funk", []byte("funk fresh"))

	c := cache{cacheDb}

	fileList, err := getFiles(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	bar := pb.StartNew(len(fileList))
	bar.SetWriter(os.Stderr)

	imgdata := make(map[string]*imageInfo)
	for _, imgpath := range fileList {
		bar.Increment()

		cName := getCacheName(imgpath)
		if cName == "" {
			continue
		}

		imginfo := c.loadCache(cName)
		if imginfo == nil {
			imginfo, err = newImageInfo(imgpath)
			if imginfo == nil {
				continue
			}
			if err != nil {
				log.Println(err)
				continue
			}

			err := c.storeCache(cName, imginfo)
			if err != nil {
				log.Fatal(err)
			}
		}

		imgdata[imgpath] = imginfo
	}

	bar.Finish()

	displayDiff(fileList, imgdata)
}

func displayDiff(fileList []string, imgdata map[string]*imageInfo) {
	fileLength := len(fileList)
	for i := 0; i < fileLength; i++ {
		for j := i + 1; j < fileLength; j++ {

			leftf := fileList[i]
			rightf := fileList[j]

			leftimg, ok1 := imgdata[leftf]
			rightimg, ok2 := imgdata[rightf]

			if ok1 && ok2 {

				avgdata1 := leftimg.Data
				avgdata2 := rightimg.Data

				if leftf == rightf {
					continue
				}

				xdiff := getDiff(avgdata1, avgdata2)

				if xdiff < uint64(*tolerance) {

					fmt.Println(leftf)
					fmt.Printf("    %d x %d\n    %s\n", leftimg.Bounds.Dx(), leftimg.Bounds.Dy(), humanize.Bytes(leftimg.Filesize))

					fmt.Println(rightf)
					fmt.Printf("    %d x %d\n    %s\n", rightimg.Bounds.Dx(), rightimg.Bounds.Dy(), humanize.Bytes(rightimg.Filesize))

					fmt.Println("")
					fmt.Println("Diff: ", xdiff)

					if xdiff > 0 && leftimg.Filesize != rightimg.Filesize {
						if *difftool != "" {
							diffTool(*difftool, leftf, rightf)
						}
					}

					fmt.Println("- - - - - - - - - -")
				}

			}

		}
	}
}

func diffTool(tool string, leftf string, rightf string) {
	log.Println("Launching difftool")
	cmd := exec.Command(tool, leftf, rightf)
	cmd.Run()
	time.Sleep(500 * time.Millisecond)
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
