package utils

import "io"

type ReadWriter struct {
	r io.Reader
	w io.Writer
}

func (rw *ReadWriter) Write(p []byte) (n int, err error) {
	return rw.w.Write(p)
}

func (rw *ReadWriter) Read(p []byte) (n int, err error) {
	return rw.r.Read(p)
}

func WrapReadWriter(r io.Reader, w io.Writer) *ReadWriter {
	return &ReadWriter{r: r, w: w}
}
