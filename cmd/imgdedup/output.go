package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/donatj/imgdedup"
	"github.com/dustin/go-humanize"
)

type ImgDiff struct {
	Left  *imgdedup.ImageInfo
	Right *imgdedup.ImageInfo

	Diff uint64
}

func diff(imgdata map[string]*imgdedup.ImageInfo, tolerance uint64) []ImgDiff {
	out := []ImgDiff{}

	fileList := make([]string, 0, len(imgdata))
	for k := range imgdata {
		fileList = append(fileList, k)
	}
	sort.Strings(fileList)

	fileLength := len(fileList)
	for i := 0; i < fileLength; i++ {
		for j := i + 1; j < fileLength; j++ {

			leftf := fileList[i]
			rightf := fileList[j]

			leftimg, ok1 := imgdata[leftf]
			rightimg, ok2 := imgdata[rightf]

			if ok1 && ok2 {

				if leftimg.Path == rightimg.Path {
					continue
				}

				xdiff, err := imgdedup.DiffImageInfos(leftimg, rightimg)
				if err != nil {
					log.Println(err)
					continue
				}

				if xdiff < tolerance {
					out = append(out, ImgDiff{
						Left:  leftimg,
						Right: rightimg,
						Diff:  xdiff,
					})
				}
			}

		}
	}

	return out
}

func displayDiff(diffs []ImgDiff) {
	for _, diff := range diffs {

		fmt.Println(diff.Left.Path)
		fmt.Printf("    %d x %d\n    %s\n", diff.Left.Bounds.Dx(), diff.Left.Bounds.Dy(), humanize.Bytes(diff.Left.Filesize))

		fmt.Println(diff.Right.Path)
		fmt.Printf("    %d x %d\n    %s\n", diff.Right.Bounds.Dx(), diff.Right.Bounds.Dy(), humanize.Bytes(diff.Right.Filesize))

		fmt.Println("")
		fmt.Println("Diff: ", diff.Diff)

		fmt.Println("- - - - - - - - - -")
	}
}