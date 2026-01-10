package lru

import (
	"reflect"
	"testing"
)

type Str string

func (d Str) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lru := New(int64(1000), nil)
	lru.Add("key1", Str("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(Str)) != "1234" {
		t.Fatal("lru cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatal("cache miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "val1", "val2", "val3"
	cap := len(k1 + v1 + k2 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, Str(v1))
	lru.Add(k2, Str(v2))
	lru.Add(k3, Str(v3))

	if lru.Len() != 2 {
		t.Fatal("lru cache length should be 2")
	}
	if _, exist := lru.Get(k1); exist {
		t.Fatal("lru cache should not get v1 successfully")
	}
	if _, exist := lru.Get(k2); !exist {
		t.Fatal("lru cache failed to get v2")
	}
	if _, exist := lru.Get(k3); !exist {
		t.Fatal("lru cache failed to get v3")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("k1", Str("123456"))
	lru.Add("k2", Str("k2"))
	lru.Add("k3", Str("k3"))
	lru.Add("k4", Str("k4"))

	expect := []string{"k1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
