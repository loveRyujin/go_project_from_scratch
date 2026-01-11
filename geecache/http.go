package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultPath = "/geecache/"

type Server struct {
	basepath string
	selfaddr string
}

func NewServer(selfaddr string) *Server {
	return &Server{
		basepath: defaultPath,
		selfaddr: selfaddr,
	}
}

func (s *Server) error(w http.ResponseWriter, code int, format string, args ...any) {
	s.log(format, args...)
	http.Error(w, fmt.Sprintf(format, args...), code)
}

func (s *Server) log(format string, args ...any) {
	log.Printf("[Server %s] %s", s.selfaddr, fmt.Sprintf(format, args...))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(v.ByteSlice())
}
