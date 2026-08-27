package cache

import (
	"image"
	"reflect"
	"testing"

	"github.com/donatj/imgdedup"
	"go.mills.io/bitcask/v2"
)

func TestCacheStoresAndLoadsImageInfo(t *testing.T) {
	db, err := bitcask.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	cache := New(db)
	imginfo := &imgdedup.ImageInfo{
		Data:     imgdedup.Pictable{{{1, 2, 3}}},
		Format:   "png",
		Bounds:   image.Rect(1, 2, 3, 4),
		Path:     "/tmp/image.png",
		Filesize: 123,
	}

	if err := cache.StoreCache("image", imginfo); err != nil {
		t.Fatal(err)
	}

	if got := cache.LoadCache("missing"); got != nil {
		t.Fatalf("LoadCache(missing) = %#v, want nil", got)
	}

	if got := cache.LoadCache("image"); !reflect.DeepEqual(got, imginfo) {
		t.Fatalf("LoadCache(image) = %#v, want %#v", got, imginfo)
	}
}
