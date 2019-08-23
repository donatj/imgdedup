package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/donatj/imgdedup"
	"github.com/donatj/imgdedup/cache"
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

	c := cache.New(cacheDb)

	fileList, err := getFiles(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	bar := pb.StartNew(len(fileList))
	bar.SetWriter(os.Stderr)

	imgdata := make(map[string]*imgdedup.ImageInfo)
	for _, imgpath := range fileList {
		bar.Increment()

		cName := cache.GetCacheName(imgpath, *subdivisions)
		if cName == "" {
			continue
		}

		imginfo := c.LoadCache(cName)
		if imginfo == nil {
			imginfo, err = imgdedup.NewImageInfo(imgpath, *subdivisions)
			if imginfo == nil {
				continue
			}
			if err != nil {
				log.Println(err)
				continue
			}

			err := c.StoreCache(cName, imginfo)
			if err != nil {
				log.Fatal(err)
			}
		}

		imgdata[imgpath] = imginfo
	}

	bar.Finish()

	displayDiff(fileList, imgdata)
}

func diffTool(tool string, leftf string, rightf string) {
	log.Println("Launching difftool")
	cmd := exec.Command(tool, leftf, rightf)
	cmd.Run()
	time.Sleep(500 * time.Millisecond)
}
