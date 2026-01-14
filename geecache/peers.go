package geecache

type PeerSeeker interface {
	Seek(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(group, key string) ([]byte, error)
}
