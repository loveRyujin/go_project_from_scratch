package consistenthash

import (
	"fmt"
	"hash/crc32"
	"slices"
	"sort"
	"strconv"
)

type Hash func([]byte) uint32

type Ring struct {
	hash     Hash
	replicas int
	hashMap  map[int]string
	keys     []int
}

func New(replicas int, hashFunc Hash) *Ring {
	if hashFunc == nil {
		hashFunc = crc32.ChecksumIEEE
	}
	return &Ring{
		hash:     hashFunc,
		replicas: replicas,
		hashMap:  make(map[int]string),
		keys:     make([]int, 0),
	}
}

func (r *Ring) Add(keys ...string) {
	for _, key := range keys {
		for i := range r.replicas {
			hash := int(r.hash([]byte(fmt.Sprintf("%s%s", strconv.Itoa(i), key))))
			r.keys = append(r.keys, hash)
			r.hashMap[hash] = key
		}
	}
	slices.Sort(r.keys)
}

func (r *Ring) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	hash := int(r.hash([]byte(key)))
	idx := sort.Search(len(r.keys), func(i int) bool {
		return r.keys[i] >= hash
	})

	return r.hashMap[r.keys[idx%len(r.keys)]]
}
