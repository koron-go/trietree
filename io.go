package trietree

import (
	"bufio"
	"encoding/binary"
	"io"
)

type writer struct {
	w   *bufio.Writer
	b   []byte
	err error
}

func newWriter(w io.Writer) *writer {
	return &writer{
		w: bufio.NewWriter(w),
		b: make([]byte, binary.MaxVarintLen64),
	}
}

func (w *writer) writeRune(v rune) {
	w.writeInt64(int64(v))
}

func (w *writer) writeInt(v int) {
	w.writeInt64(int64(v))
}

func (w *writer) writeInt64(v int64) error {
	if w.err != nil {
		return w.err
	}
	n := binary.PutVarint(w.b, v)
	_, err := w.w.Write(w.b[:n])
	if err != nil {
		w.err = err
		return err
	}
	return nil
}

type wrapByteReader struct {
	io.Reader
}

func (br wrapByteReader) ReadByte() (byte, error) {
	var b [1]byte
	n, err := br.Read(b[:])
	if n != 1 {
		return 0, err
	}
	return b[0], nil
}

func toByteReader(r io.Reader) io.ByteReader {
	if br, ok := r.(io.ByteReader); ok {
		return br
	}
	return wrapByteReader{r}
}

type reader struct {
	r   io.ByteReader
	err error
}

func newReader(r io.Reader) *reader {
	return &reader{
		r: toByteReader(r),
	}
}

func (r *reader) readRune() rune {
	n, _ := r.readInt64()
	return rune(n)
}

func (r *reader) readInt() int {
	n, _ := r.readInt64()
	return int(n)
}

func (r *reader) readInt64() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	n, err := binary.ReadVarint(r.r)
	if err != nil {
		r.err = err
		return 0, err
	}
	return n, nil
}
