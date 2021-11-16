package utils

import (
	"bufio"
	"io"
)

// Wraps a bufio reader and writer in one object.
// The bufio package provides already a bufio.ReadWriter which also combines the functionalities
// of a bufio Reader and a Writer, but with the great drawback that the buffering is double layered,
// so flushing the written content to a bufio.ReadWriter will write it's content to the underlying bufio.Writer
// which buffers it again and does NOT flush it forward to the "real" writer (file/connection etc).
type BufferedReadWriter struct {
	bufr *bufio.Reader
	bufw *bufio.Writer
	r    io.Reader
	w    io.Writer
}

func NewBufferedReadWriter(r io.Reader, w io.Writer) *BufferedReadWriter {
	return &BufferedReadWriter{bufr: bufio.NewReader(r), bufw: bufio.NewWriter(w), r: r, w: w}
}

func (rw *BufferedReadWriter) Write(p []byte) (n int, err error) {
	return rw.bufw.Write(p)
}

func (rw *BufferedReadWriter) Flush() error {
	return rw.bufw.Flush()
}

func (rw *BufferedReadWriter) Read(p []byte) (n int, err error) {
	return rw.bufr.Read(p)
}

func (rw *BufferedReadWriter) ReadString(delim byte) (string, error) {
	return rw.bufr.ReadString(delim)
}

func (rw *BufferedReadWriter) Reset() {
	rw.bufw.Reset(rw.w)
	rw.bufr.Reset(rw.r)
}
