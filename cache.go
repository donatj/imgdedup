package main

import (
	"bufio"
	"encoding/json"
	"os"
)

func loadCache(cachename string) (*imageInfo, error) {
	file, err := os.Open(cachename)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(file)

	dec := json.NewDecoder(r)

	var imginfo imageInfo
	err = dec.Decode(&imginfo)
	if err != nil {
		return nil, err
	}

	return &imginfo, nil
}

func storeCache(cachename string, imginfo *imageInfo) {
	fo, err := os.Create(cachename)
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	enc := json.NewEncoder(fo)

	err = enc.Encode(imginfo)
	if err != nil {
		panic(err)
	}

}
