package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/loveRyujin/geecache"
)

var db = map[string]string{
	"test1": "val1",
	"test2": "val2",
	"test3": "val3",
}

func main() {
	geecache.NewGroup("test_group", 2<<10, geecache.LoadFunc(func(key string) ([]byte, error) {
		log.Printf("[slow DB] search key: %s", key)
		v, exist := db[key]
		if !exist {
			return nil, fmt.Errorf("%s is not existed", key)
		}

		return []byte(v), nil
	}))

	addr := "localhost:8431"
	server := geecache.NewServer(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, server))
}
