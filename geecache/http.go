package geecache

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/loveRyujin/geecache/consistenthash"
	"github.com/loveRyujin/geecache/geecachepb"
	"google.golang.org/protobuf/proto"
)

const (
	defaultPath     = "/geecache/"
	defaultReplicas = 10
)

type HTTPPool struct {
	basepath    string
	selfaddr    string
	mu          sync.Mutex
	peers       *consistenthash.Ring
	httpGetters map[string]*httpGetter
}

var _ PeerSeeker = (*HTTPPool)(nil)

func NewHTTPPool(selfaddr string) *HTTPPool {
	return &HTTPPool{
		basepath: defaultPath,
		selfaddr: selfaddr,
	}
}

func (s *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.log("Method: %s, Path: %s", r.Method, r.URL.Path)

	// request example:
	// http://xxx.xxx.xxx.xxx:80/geecache/test_group/test1
	if !strings.HasPrefix(r.URL.Path, s.basepath) {
		s.error(w, http.StatusBadRequest, "invalid request path: %s", r.URL.Path)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, s.basepath)
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		s.error(w, http.StatusBadRequest, "invalid request path: %s", r.URL.Path)
		return
	}

	groupName := parts[0]
	cacheKey := parts[1]

	g := GetGroup(groupName)
	if g == nil {
		s.error(w, http.StatusNotFound, "cache group not found: %s", groupName)
		return
	}

	v, err := g.Get(cacheKey)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "server error: %v", err.Error())
		return
	}

	data, err := proto.Marshal(&geecachepb.Response{Value: v.ByteSlice()})
	if err != nil {
		s.error(w, http.StatusInternalServerError, "proto marshal error: %v", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func (s *HTTPPool) Set(peers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.peers = consistenthash.New(defaultReplicas, nil)
	s.peers.Add(peers...)
	s.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		s.httpGetters[peer] = &httpGetter{
			baseURL: peer + s.basepath,
			client: &http.Client{
				Timeout: 5 * time.Second, // 设置5秒超时
			},
		}
	}
}

func (s *HTTPPool) Seek(key string) (PeerGetter, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peer := s.peers.Get(key)
	if peer == "" || peer == s.selfaddr {
		return nil, false
	}

	s.log("Pick peer %s", peer)
	return s.httpGetters[peer], true
}

func (s *HTTPPool) error(w http.ResponseWriter, code int, format string, args ...any) {
	s.log(format, args...)
	http.Error(w, fmt.Sprintf(format, args...), code)
}

func (s *HTTPPool) log(format string, args ...any) {
	log.Printf("[Server %s] %s", s.selfaddr, fmt.Sprintf(format, args...))
}

type httpGetter struct {
	baseURL string
	client  *http.Client
}

var _ PeerGetter = (*httpGetter)(nil)

func (hg *httpGetter) Get(in *geecachepb.Request) (*geecachepb.Response, error) {
	// example: http://localhost:8080/geecache/{:group}/{:key}
	url := fmt.Sprintf("%v%v/%v", hg.baseURL, url.QueryEscape(in.Group), url.QueryEscape(in.Key))
	log.Println(url)
	resp, err := hg.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	out := &geecachepb.Response{}
	if err = proto.Unmarshal(bytes, out); err != nil {
		return nil, fmt.Errorf("decoding response body: %v", err)
	}

	return out, nil
}
