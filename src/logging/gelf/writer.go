package gelf

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/rand"
	"fmt"
	"io"
	"net"
)

// implement io.WriteCloser.
type writeCloser struct {
	*bytes.Buffer
}

// implement io.Writer
type writer struct {
	conn             net.Conn
	chunkSize        int
	chunkDataSize    int
	compressionType  int
	compressionLevel int
	protocol         string
}

func (w *writer) Write(buf []byte) (n int, err error) {
	var (
		cw   io.WriteCloser
		cBuf bytes.Buffer
	)

	if w.protocol == "tcp" {
		cBuf.Write(buf)
		cBuf.Write([]byte("\000"))

		if _, err := w.conn.Write(cBuf.Bytes()); err != nil {
			return n, err
		}
		if err != nil {
			return 0, err
		}

		return n, nil
	}

	switch w.compressionType {
	case CompressionNone:
		cw = &writeCloser{&cBuf}
	case CompressionGzip:
		cw, err = gzip.NewWriterLevel(&cBuf, w.compressionLevel)
	case CompressionZlib:
		cw, err = zlib.NewWriterLevel(&cBuf, w.compressionLevel)
	}

	if n, err = cw.Write(buf); err != nil {
		return n, err
	}

	cw.Close()

	var cBytes = cBuf.Bytes()
	if count := w.chunkCount(cBytes); count > 1 {
		return w.writeChunked(count, cBytes)
	}

	if n, err = w.conn.Write(cBytes); err != nil {
		return n, err
	}

	if n != len(cBytes) {
		return n, fmt.Errorf("writed %d bytes but should %d bytes", n, len(cBytes))
	}

	return n, nil
}

// Close implementation of io.WriteCloser.
func (*writeCloser) Close() error {
	return nil
}

// chunkCount calculate the number of GELF chunks.
func (w *writer) chunkCount(b []byte) int {
	lenB := len(b)
	if lenB <= w.chunkSize {
		return 1
	}

	return len(b)/w.chunkDataSize + 1
}

func (w *writer) writeChunked(count int, cBytes []byte) (n int, err error) {
	if count > MaxChunkCount {
		return 0, fmt.Errorf("need %d chunks but shold be later or equal to %d", count, MaxChunkCount)
	}

	var (
		cBuf = bytes.NewBuffer(
			make([]byte, 0, w.chunkSize),
		)
		nChunks   = uint8(count)
		messageID = make([]byte, 8)
	)

	if n, err = io.ReadFull(rand.Reader, messageID); err != nil || n != 8 {
		return 0, fmt.Errorf("rand.Reader: %d/%s", n, err)
	}

	var (
		off       int
		chunkLen  int
		bytesLeft = len(cBytes)
	)

	for i := uint8(0); i < nChunks; i++ {
		off = int(i) * w.chunkDataSize
		chunkLen = w.chunkDataSize
		if chunkLen > bytesLeft {
			chunkLen = bytesLeft
		}

		cBuf.Reset()
		cBuf.Write(chunkedMagicBytes)
		cBuf.Write(messageID)
		cBuf.WriteByte(i)
		cBuf.WriteByte(nChunks)
		cBuf.Write(cBytes[off : off+chunkLen])

		if n, err = w.conn.Write(cBuf.Bytes()); err != nil {
			return len(cBytes) - bytesLeft + n, err
		}

		if n != len(cBuf.Bytes()) {
			n = len(cBytes) - bytesLeft + n
			return n, fmt.Errorf("writed %d bytes but should %d bytes", n, len(cBytes))
		}

		bytesLeft -= chunkLen
	}

	if bytesLeft != 0 {
		return len(cBytes) - bytesLeft, fmt.Errorf("error: %d bytes left after sending", bytesLeft)
	}

	return len(cBytes), nil
}
