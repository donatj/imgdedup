package cache

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"sync"

	"git.mills.io/prologic/bitcask"
	"github.com/donatj/imgdedup"
)

type Cache struct {
	db *bitcask.Bitcask

	sync.Mutex
}

func New(db *bitcask.Bitcask) *Cache {
	return &Cache{
		db: db,
	}
}

func (c *Cache) LoadCache(cachename string) *imgdedup.ImageInfo {
	c.Lock()
	defer c.Unlock()

	b, err := c.db.Get([]byte(cachename))
	if err == bitcask.ErrKeyNotFound {
		return nil
	} else if err != nil {
		panic(err)
	}

	data := bytes.NewBuffer(b)
	dec := gob.NewDecoder(data)

	imginfo := imgdedup.ImageInfo{}
	err = dec.Decode(&imginfo)
	if err != nil {
		return nil
	}

	return &imginfo
}

func (c *Cache) StoreCache(cachename string, imginfo *imgdedup.ImageInfo) error {
	c.Lock()
	defer c.Unlock()

	data := &bytes.Buffer{}

	enc := gob.NewEncoder(data)
	err := enc.Encode(*imginfo)
	if err != nil {
		return err
	}

	err = c.db.Put([]byte(cachename), data.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func GetCacheName(imgpath string, subdivisions uint) string {
	file, err := os.Open(imgpath)
	if err != nil {
		return ""
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return ""
	}

	str := fmt.Sprintf("%s|%d|%d|%d", imgpath, subdivisions, fi.Size(), fi.ModTime().Unix())

	h := md5.New()
	io.WriteString(h, str)

	return fmt.Sprintf("%x", h.Sum(nil))
}
