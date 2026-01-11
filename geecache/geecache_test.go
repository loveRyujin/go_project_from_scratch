package geecache

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"test1": "val1",
	"test2": "val2",
	"test3": "val3",
}

func TestGeeCacheGet(t *testing.T) {
	countM := make(map[string]int, len(db))
	geecache := NewGroup("test_group", 2<<10, LoadFunc(func(key string) ([]byte, error) {
		log.Printf("[slow DB] search key: %s", key)
		v, exist := db[key]
		if !exist {
			return nil, fmt.Errorf("%s is not existed", key)
		}

		countM[key] += 1
		return []byte(v), nil
	}))
	if g, ok := groups["test_group"]; !ok || g != geecache {
		t.Fatal("Failed to get cache group")
	}

	for k, v := range db {
		if res, err := geecache.Get(k); err != nil || res.String() != v {
			t.Fatalf("Failed to get key: %s", k)
		}
		if _, err := geecache.Get(k); err != nil || countM[k] > 1 {
			t.Fatalf("Failed to cache key: %s", k)
		}
	}
}
