package main

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"io"
	"os"

	"github.com/prologic/bitcask"
)

type cache struct {
	db *bitcask.Bitcask
}

func (c *cache) loadCache(cachename string) *imageInfo {
	b, err := c.db.Get(cachename)
	if err == bitcask.ErrKeyNotFound {
		return nil
	} else if err != nil {
		panic(err)
	}

	data := bytes.NewBuffer(b)
	dec := gob.NewDecoder(data)

	imginfo := imageInfo{}
	err = dec.Decode(&imginfo)
	if err != nil {
		return nil
	}

	return &imginfo
}

func (c *cache) storeCache(cachename string, imginfo *imageInfo) error {
	data := &bytes.Buffer{}

	enc := gob.NewEncoder(data)
	err := enc.Encode(*imginfo)
	if err != nil {
		return err
	}

	err = c.db.Put(cachename, data.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func getCacheName(imgpath string) string {
	file, err := os.Open(imgpath)
	if err != nil {
		return ""
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return ""
	}

	str := imgpath + "|" + string(*subdivisions) + "|" + string(fi.Size()) + string(fi.ModTime().Unix())

	h := md5.New()
	io.WriteString(h, str)

	return fmt.Sprintf("%x", h.Sum(nil))
}
