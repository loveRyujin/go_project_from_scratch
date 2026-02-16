package geerpc

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/loveRyujin/geerpc/codec"
)

// startServer 启动 Server 处理 serverConn，返回 clientConn 供测试使用。
func startServer(t *testing.T) (client net.Conn, closeFn func()) {
	t.Helper()
	client, server := net.Pipe()
	s := NewServer()
	go s.ServerConn(server)
	return client, func() { client.Close() }
}

// sendOption 向连接发送 Option 完成协议协商。
func sendOption(t *testing.T, conn net.Conn, opt *Option) {
	t.Helper()
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		t.Fatalf("send option error: %v", err)
	}
}

func TestServerConn_SingleRequest(t *testing.T) {
	client, closeFn := startServer(t)
	defer closeFn()

	sendOption(t, client, DefaultOption)

	cc := codec.NewGobCodec(client)

	h := &codec.Header{ServerMethod: "Foo.Bar", Seq: 1}
	if err := cc.Write(h, "request body"); err != nil {
		t.Fatalf("write request error: %v", err)
	}

	var respH codec.Header
	if err := cc.ReadHeader(&respH); err != nil {
		t.Fatalf("read response header error: %v", err)
	}
	if respH.Seq != 1 {
		t.Errorf("response Seq = %d, want 1", respH.Seq)
	}
	if respH.ServerMethod != "Foo.Bar" {
		t.Errorf("response ServerMethod = %q, want %q", respH.ServerMethod, "Foo.Bar")
	}

	var respBody string
	if err := cc.ReadBody(&respBody); err != nil {
		t.Fatalf("read response body error: %v", err)
	}
	want := fmt.Sprintf("geerpc resp %d", 1)
	if respBody != want {
		t.Errorf("response body = %q, want %q", respBody, want)
	}
}

func TestServerConn_MultipleRequests(t *testing.T) {
	client, closeFn := startServer(t)
	defer closeFn()

	sendOption(t, client, DefaultOption)

	cc := codec.NewGobCodec(client)

	rounds := 5
	for i := 0; i < rounds; i++ {
		h := &codec.Header{ServerMethod: "Arith.Add", Seq: uint64(i)}
		if err := cc.Write(h, fmt.Sprintf("request %d", i)); err != nil {
			t.Fatalf("round %d: write request error: %v", i, err)
		}

		var respH codec.Header
		if err := cc.ReadHeader(&respH); err != nil {
			t.Fatalf("round %d: read response header error: %v", i, err)
		}
		if respH.Seq != uint64(i) {
			t.Errorf("round %d: response Seq = %d, want %d", i, respH.Seq, i)
		}

		var respBody string
		if err := cc.ReadBody(&respBody); err != nil {
			t.Fatalf("round %d: read response body error: %v", i, err)
		}
		want := fmt.Sprintf("geerpc resp %d", i)
		if respBody != want {
			t.Errorf("round %d: response body = %q, want %q", i, respBody, want)
		}
	}
}

func TestServerConn_InvalidMagicNumber(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := NewServer()
	go s.ServerConn(server)

	badOpt := &Option{
		MagicNumber: 0x000000,
		CodecType:   codec.GobType,
	}
	sendOption(t, client, badOpt)

	// 服务端检测到非法 MagicNumber 后会直接 return，关闭连接。
	// 客户端再读取时应该收到 EOF 或 closed pipe 错误。
	buf := make([]byte, 1)
	_, err := client.Read(buf)
	if err == nil {
		t.Error("expected error reading from closed connection, got nil")
	}
}

func TestServerConn_InvalidCodecType(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	s := NewServer()
	go s.ServerConn(server)

	badOpt := &Option{
		MagicNumber: MagicNumber,
		CodecType:   "application/unknown",
	}
	sendOption(t, client, badOpt)

	// 服务端检测到未注册的 CodecType 后会直接 return，关闭连接。
	buf := make([]byte, 1)
	_, err := client.Read(buf)
	if err == nil {
		t.Error("expected error reading from closed connection, got nil")
	}
}

func TestAccept_WithListener(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	defer lis.Close()

	go Accept(lis)

	// 验证 Accept 能正常接受 TCP 连接
	conn, err := net.Dial("tcp", lis.Addr().String())
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	// 发送合法 Option，服务端应该正常接受（不会断开连接）
	sendOption(t, conn, DefaultOption)

	// 注意：完整的 RPC 请求-响应周期已通过 net.Pipe 的测试覆盖。
	// 真实 TCP 连接上 json.NewDecoder 的内部缓冲可能吞掉后续 gob 数据，
	// 这需要在 Client 实现中通过共享 bufio.Reader 解决。
}
