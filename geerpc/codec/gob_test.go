package codec

import (
	"bytes"
	"io"
	"sync"
	"testing"
)

// mockConn 用 io.Pipe 模拟一个全双工的网络连接，
// 写入端和读取端配对：一端 Write 的数据，另一端可以 Read 到。
type mockConn struct {
	r io.Reader
	w io.Writer
	io.Closer
}

func (m *mockConn) Read(p []byte) (int, error)  { return m.r.Read(p) }
func (m *mockConn) Write(p []byte) (int, error) { return m.w.Write(p) }

// newMockConnPair 创建一对连接，模拟 client ↔ server 的双向通信。
// clientConn 写入的数据可以从 serverConn 读取，反之亦然。
func newMockConnPair() (clientConn, serverConn io.ReadWriteCloser) {
	// client → server 方向的管道
	c2sReader, c2sWriter := io.Pipe()
	// server → client 方向的管道
	s2cReader, s2cWriter := io.Pipe()

	clientConn = &mockConn{r: s2cReader, w: c2sWriter, Closer: c2sWriter}
	serverConn = &mockConn{r: c2sReader, w: s2cWriter, Closer: s2cWriter}
	return
}

func TestGobCodec_WriteAndRead(t *testing.T) {
	clientConn, serverConn := newMockConnPair()
	defer clientConn.Close()
	defer serverConn.Close()

	writer := NewGobCodec(clientConn)
	reader := NewGobCodec(serverConn)

	h := &Header{
		ServerMethod: "Arith.Add",
		Seq:          1,
	}
	body := "geerpc request body"

	var wg sync.WaitGroup
	wg.Add(1)

	// writer 端写入
	go func() {
		defer wg.Done()
		if err := writer.Write(h, body); err != nil {
			t.Errorf("Write error: %v", err)
		}
	}()

	// reader 端读取
	var readH Header
	if err := reader.ReadHeader(&readH); err != nil {
		t.Fatalf("ReadHeader error: %v", err)
	}
	if readH.ServerMethod != h.ServerMethod {
		t.Errorf("ServerMethod = %q, want %q", readH.ServerMethod, h.ServerMethod)
	}
	if readH.Seq != h.Seq {
		t.Errorf("Seq = %d, want %d", readH.Seq, h.Seq)
	}

	var readBody string
	if err := reader.ReadBody(&readBody); err != nil {
		t.Fatalf("ReadBody error: %v", err)
	}
	if readBody != body {
		t.Errorf("Body = %q, want %q", readBody, body)
	}

	wg.Wait()
}

func TestGobCodec_MultipleRounds(t *testing.T) {
	clientConn, serverConn := newMockConnPair()
	defer clientConn.Close()
	defer serverConn.Close()

	writer := NewGobCodec(clientConn)
	reader := NewGobCodec(serverConn)

	rounds := 5
	var wg sync.WaitGroup
	wg.Add(1)

	// writer 端连续写入多轮
	go func() {
		defer wg.Done()
		for i := range rounds {
			h := &Header{
				ServerMethod: "Arith.Multiply",
				Seq:          uint64(i),
			}
			body := i * 100
			if err := writer.Write(h, body); err != nil {
				t.Errorf("round %d: Write error: %v", i, err)
				return
			}
		}
	}()

	// reader 端连续读取多轮
	for i := range rounds {
		var readH Header
		if err := reader.ReadHeader(&readH); err != nil {
			t.Fatalf("round %d: ReadHeader error: %v", i, err)
		}
		if readH.Seq != uint64(i) {
			t.Errorf("round %d: Seq = %d, want %d", i, readH.Seq, i)
		}
		if readH.ServerMethod != "Arith.Multiply" {
			t.Errorf("round %d: ServerMethod = %q, want %q", i, readH.ServerMethod, "Arith.Multiply")
		}

		var readBody int
		if err := reader.ReadBody(&readBody); err != nil {
			t.Fatalf("round %d: ReadBody error: %v", i, err)
		}
		if readBody != i*100 {
			t.Errorf("round %d: Body = %d, want %d", i, readBody, i*100)
		}
	}

	wg.Wait()
}

func TestGobCodec_StructBody(t *testing.T) {
	clientConn, serverConn := newMockConnPair()
	defer clientConn.Close()
	defer serverConn.Close()

	writer := NewGobCodec(clientConn)
	reader := NewGobCodec(serverConn)

	type Args struct {
		A, B int
	}

	h := &Header{ServerMethod: "Arith.Add", Seq: 42}
	body := Args{A: 3, B: 5}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := writer.Write(h, body); err != nil {
			t.Errorf("Write error: %v", err)
		}
	}()

	var readH Header
	if err := reader.ReadHeader(&readH); err != nil {
		t.Fatalf("ReadHeader error: %v", err)
	}
	if readH.Seq != 42 {
		t.Errorf("Seq = %d, want 42", readH.Seq)
	}

	var readBody Args
	if err := reader.ReadBody(&readBody); err != nil {
		t.Fatalf("ReadBody error: %v", err)
	}
	if readBody.A != 3 || readBody.B != 5 {
		t.Errorf("Body = %+v, want {A:3, B:5}", readBody)
	}

	wg.Wait()
}

func TestNewCodecFuncMap(t *testing.T) {
	// GobType 应该已注册
	if fn := NewCodecFuncMap[GobType]; fn == nil {
		t.Fatal("GobType not registered in NewCodecFuncMap")
	}

	// JsonType 尚未注册
	if fn := NewCodecFuncMap[JsonType]; fn != nil {
		t.Fatal("JsonType should not be registered yet")
	}

	// 通过注册表创建的 Codec 应该可以正常工作
	var buf bytes.Buffer
	conn := &nopCloserRWC{Buffer: &buf}
	codec := NewCodecFuncMap[GobType](conn)
	if codec == nil {
		t.Fatal("NewCodecFuncMap[GobType] returned nil Codec")
	}
}

func TestGobCodec_Close(t *testing.T) {
	clientConn, serverConn := newMockConnPair()
	defer serverConn.Close()

	codec := NewGobCodec(clientConn)

	// 第一次 Close 应该成功
	if err := codec.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// 关闭后，对端读取应该收到 EOF
	buf := make([]byte, 1)
	_, err := serverConn.Read(buf)
	if err != io.EOF && err != io.ErrClosedPipe {
		t.Errorf("Read from closed peer: got err = %v, want EOF or ErrClosedPipe", err)
	}
}

// nopCloserRWC 把 bytes.Buffer 包装成 io.ReadWriteCloser
type nopCloserRWC struct {
	*bytes.Buffer
}

func (n *nopCloserRWC) Close() error { return nil }
