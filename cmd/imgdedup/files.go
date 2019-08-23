package main

import (
	"os"
	"path/filepath"
	"strings"
)

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

func filterFiles(paths []string, exts []string) []string {
	n := 0
pathLoop:
	for _, path := range paths {
		fExt := strings.ToLower(filepath.Ext(path))
		for _, ext := range exts {
			if fExt == ext {
				paths[n] = path
				n++
				continue pathLoop
			}
		}
	}
	return paths[:n]
}
