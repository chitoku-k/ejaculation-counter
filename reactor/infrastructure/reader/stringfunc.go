//go:generate go tool mockgen -source=stringfunc.go -destination=stringfunc_mock.go -package=reader -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/reader

package reader

import (
	"io"
)

type StringGenerator interface {
	Generate() string
}

type stringFuncReader struct {
	buf       []byte
	separator string
	current   int
	length    int
	gen       func() string
}

func NewStringFuncReader(separator string, length int, gen func() string) io.Reader {
	return &stringFuncReader{
		separator: separator,
		length:    length,
		gen:       gen,
	}
}

func (s *stringFuncReader) Read(p []byte) (n int, err error) {
	n = copy(p, s.buf)
	s.buf = s.buf[n:]

	if n > 0 {
		return
	}

	if s.current == s.length {
		err = io.EOF
		return
	}

	var str string
	if s.current > 0 {
		str = s.separator
	}

	str += s.gen()
	n += copy(p[n:], str)
	s.buf = []byte(str)[n:]
	s.current++
	return
}
