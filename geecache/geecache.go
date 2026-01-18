package geecache

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/loveRyujin/geecache/geecachepb"
	"github.com/loveRyujin/geecache/singleflight"
)

type Loader interface {
	Load(key string) ([]byte, error)
}

type LoadFunc func(key string) ([]byte, error)

func (lf LoadFunc) Load(key string) ([]byte, error) {
	return lf(key)
}

var (
	rwmu   sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	maincache    *cache
	loader       Loader
	name         string
	peers        PeerSeeker
	singleflight *singleflight.Group
}

func NewGroup(name string, cacheBytes int64, loader Loader) *Group {
	if loader == nil {
		panic("loader is nil")
	}

	g := &Group{
		name:         name,
		maincache:    &cache{cacheBytes: cacheBytes},
		loader:       loader,
		singleflight: &singleflight.Group{},
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

func (g *Group) RegisterPeers(peers PeerSeeker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
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
	val, err := g.singleflight.Do(key, func() (any, error) {
		if g.peers != nil {
			p, ok := g.peers.Seek(key)
			if ok {
				res, err := g.getFromPeer(p, key)
				if err == nil {
					return res, nil
				}
				log.Println("[GeeCache] Failed to get from peer:", err)
			}
		}

		return g.getLocally(key)
	})
	if err != nil {
		return ByteView{}, err
	}

	return val.(ByteView), nil
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	if peer == nil {
		return ByteView{}, fmt.Errorf("peer is empty")
	}

	in := &geecachepb.Request{
		Group: g.name,
		Key:   key,
	}
	res, err := peer.Get(in)
	if err != nil {
		return ByteView{}, err
	}

	return ByteView{b: res.Value}, nil
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
