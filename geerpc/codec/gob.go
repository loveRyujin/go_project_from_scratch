package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn    io.ReadWriteCloser
	buf     *bufio.Writer
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn:    conn,
		buf:     buf,
		encoder: gob.NewEncoder(buf),
		decoder: gob.NewDecoder(conn),
	}
}

func (gc *GobCodec) ReadHeader(h *Header) error {
	return gc.decoder.Decode(h)
}

func (gc *GobCodec) ReadBody(body any) error {
	return gc.decoder.Decode(body)
}

func (gc *GobCodec) Write(h *Header, body any) error {
	var err error
	defer func() {
		_ = gc.buf.Flush()
		if err != nil {
			_ = gc.Close()
		}
	}()

	if err = gc.encoder.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header", err)
		return err
	}
	if err = gc.encoder.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body", err)
		return err
	}

	return nil
}

func (gc *GobCodec) Close() error {
	return gc.conn.Close()
}
