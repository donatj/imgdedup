package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"git.mills.io/prologic/bitcask"
	"github.com/donatj/imgdedup"
	"github.com/donatj/imgdedup/cache"
	"github.com/schollz/progressbar/v3"

	// Image format self registers

	// Standard Image Formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	// Extended Image Formats
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var (
	subdivisions = flag.Uint("subdivisions", 10, "Slices per axis")
	tolerance    = flag.Uint64("tolerance", 100, "Color delta tolerance, higher = more tolerant")
	format       = flag.String("format", "default", "Output format - available options: default classic json")
	difftool     = flag.String("diff", "", "Command to pass dupe images to eg: cmd $left $right")
	noprogress   = flag.Bool("no-progress", false, "Disable progress bar")
)

var (
	cacheDir *string
	cacheDb  *bitcask.Bitcask
)

func init() {
	cd, err := os.UserCacheDir()
	if err != nil {
		h, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		cd = filepath.Join(h, ".cache")
	}

	cacheDir = flag.String("cache-dir", filepath.Join(cd, "imgdedup", "cacheDb2"), "")
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

	c := cache.New(cacheDb)

	fileList, err := getFiles(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	fileList = filterFiles(fileList,
		[]string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".tiff"})

	// bar := progressbar.Default(int64(len(fileList)))
	var bar *progressbar.ProgressBar
	if *noprogress {
		bar = progressbar.DefaultSilent(int64(len(fileList)))
	} else {
		bar = progressbar.NewOptions(len(fileList),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionShowCount(),
		)
	}

	fileChan := make(chan string)

	imgdata := make(map[string]*imgdedup.ImageInfo)
	imgmut := sync.Mutex{}

	wg := sync.WaitGroup{}
	for i := 0; i <= runtime.NumCPU()*2; i++ {
		wg.Add(1)
		go func() {
			for {
				imgpath, ok := <-fileChan
				if !ok {
					break
				}

				cName := cache.GetCacheName(imgpath, *subdivisions)
				if cName == "" {
					log.Println("failed to handle", imgpath)
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

				imgmut.Lock()
				imgdata[imgpath] = imginfo
				imgmut.Unlock()

				bar.Add(1)
			}

			wg.Done()
		}()
	}

	for _, imgpath := range fileList {
		fileChan <- imgpath
	}
	close(fileChan)
	wg.Wait()

	bar.Finish()

	d := diff(imgdata, *tolerance)

	switch *format {
	case "default":
		displayHumanDiff(d)
	case "classic":
		displayClassicDiff(d)
	case "json":
		displayDiffJSON(d)
	default:
		log.Fatal("unhandled format", *format)
	}

	if *difftool != "" {
		diffToolDiff(*difftool, d)
	}

}
