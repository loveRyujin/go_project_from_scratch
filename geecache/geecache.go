package geecache

import (
	"errors"
	"log"
	"sync"
)

type Loader interface {
	Load(key string) ([]byte, error)
}

type LoadFunc func(key string) ([]byte, error)

func (lf LoadFunc) Load(key string) ([]byte, error) {
	return lf.Load(key)
}

var (
	rwmu   sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	maincache *cache
	loader    Loader
	name      string
}

func NewGroup(name string, cacheBytes int64, loader Loader) *Group {
	if loader == nil {
		panic("loader is nil")
	}

	g := &Group{
		name:      name,
		maincache: &cache{cacheBytes: cacheBytes},
		loader:    loader,
	}
	rwmu.Lock()
	groups[name] = g
	rwmu.Unlock()

	return g
}

func GetGroup(name string) *Group {
	rwmu.RLock()
	defer rwmu.RUnlock()

	g := groups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is empty")
	}

	if val, exist := g.maincache.Get(key); exist {
		log.Println("[GeeCache] hit")
		return val, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	res, err := g.loader.Load(key)
	if err != nil {
		return ByteView{}, err
	}

	bv := ByteView{b: cloneBytes(res)}
	g.populateCache(key, bv)

	return bv, nil
}

func (g *Group) populateCache(key string, val ByteView) {
	g.maincache.Add(key, val)
}
