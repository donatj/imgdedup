package main

import (
	"bufio"
	"encoding/json"
	"os"
)

func loadCache(cachename string) (pictable, error) {

	file, err := os.Open(cachename)
	defer file.Close()
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
	defer fo.Close()
	if err != nil {
		panic(err)
	}

	enc := json.NewEncoder(fo)
	enc.Encode(avgdata)
}