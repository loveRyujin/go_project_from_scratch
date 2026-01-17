package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/loveRyujin/geecache"
)

var (
	peerAddrs = flag.String("test_peer_addrs", "", "Comma-separated list of peer addresses")
	peerIndex = flag.Int("test_peer_index", -1, "Index of which peer this child is")
	peerChild = flag.Bool("test_peer_child", false, "True if running as a child process")
)

var db = map[string]string{
	"test1": "val1",
	"test2": "val2",
	"test3": "val3",
}

func main() {
	flag.Parse()

	if *peerChild {
		childProcess()
		os.Exit(0)
	}

	const (
		nChild = 3
		nGets  = 100
	)
	addrs := []string{
		"http://localhost:9091",
		"http://localhost:9092",
		"http://localhost:9093",
	}

	var cmds []*exec.Cmd
	var wg sync.WaitGroup
	for i := range nChild {
		cmd := exec.Command(os.Args[0],
			"--test_peer_child",
			"--test_peer_addrs="+strings.Join(addrs, ","),
			"--test_peer_index="+strconv.Itoa(i),
		)
		// 将子进程的标准输出和标准错误重定向到父进程，以便看到日志
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmds = append(cmds, cmd)
		wg.Add(1)
		if err := cmd.Start(); err != nil {
			panic("failed to start child process: " + err.Error())
		}
		go waitAddrReady(addrs[i], &wg)
	}
	defer func() {
		for i := range nChild {
			if cmds[i].Process != nil {
				cmds[i].Process.Kill()
			}
		}
	}()
	wg.Wait()

	// 等待中断信号，优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.Println("All peers started. Press Ctrl+C to exit.")
	<-sigChan
	log.Println("Shutting down...")
}

func childProcess() {
	addrs := strings.Split(*peerAddrs, ",")
	selfAddr := addrs[*peerIndex]
	p := geecache.NewHTTPPool(selfAddr)
	p.Set(addrs...)
	g := geecache.NewGroup("test_group", 2<<10, geecache.LoadFunc(func(key string) ([]byte, error) {
		log.Printf("[slow DB] search key: %s, peer ip is %s", key, selfAddr)
		v, exist := db[key]
		if !exist {
			return nil, fmt.Errorf("%s is not existed", key)
		}

		return []byte(v), nil
	}))
	g.RegisterPeers(p)

	// 从 URL 中提取 host:port
	serverAddr := extractHostPort(selfAddr)

	// 创建带超时配置的 HTTP 服务器，防止连接泄漏
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      p,
		ReadTimeout:  10 * time.Second,  // 读取超时
		WriteTimeout: 10 * time.Second,  // 写入超时
		IdleTimeout:  120 * time.Second, // 空闲连接超时
	}

	// 监听信号，优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Printf("peer %d shutting down...", *peerIndex)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("peer %d shutdown error: %v", *peerIndex, err)
		}
	}()

	log.Printf("peer %d is running on %s", *peerIndex, serverAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("peer %d failed to start: %v", *peerIndex, err)
	}
}

func waitAddrReady(addr string, wg *sync.WaitGroup) {
	defer wg.Done()

	const (
		maxDelay   = 1 * time.Second
		maxRetries = 40 // 最多重试40次，约1秒后达到最大延迟
		timeout    = 2 * time.Second
	)

	// 从 URL 中提取 host:port 用于 TCP 连接
	hostPort := extractHostPort(addr)

	tries := 0
	for tries < maxRetries {
		tries++
		conn, err := net.DialTimeout("tcp", hostPort, timeout)
		if err == nil {
			conn.Close()
			return
		}
		delay := time.Duration(tries) * 25 * time.Millisecond
		if delay > maxDelay {
			delay = maxDelay
		}
		time.Sleep(delay)
	}
	log.Printf("Failed to connect to %s after %d retries", hostPort, maxRetries)
}

// extractHostPort 从 URL 中提取 host:port
func extractHostPort(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		// 如果不是有效的 URL，尝试直接使用（可能是 host:port 格式）
		return urlStr
	}
	if u.Port() == "" {
		// 如果没有端口，使用默认端口
		if u.Scheme == "https" {
			return u.Host + ":443"
		}
		return u.Host + ":80"
	}
	return u.Host
}
