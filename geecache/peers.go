package geecache

import "github.com/loveRyujin/geecache/geecachepb"

type PeerSeeker interface {
	Seek(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(in *geecachepb.Request) (out *geecachepb.Response, err error)
}
