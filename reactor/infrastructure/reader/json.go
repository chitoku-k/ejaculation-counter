package reader

import (
	"encoding/json/jsontext"
	"io"
)

type jsonStreamReader struct {
	buf           []byte
	separator     string
	depth         int
	currentDepth  int
	currentLength int
	decoder       *jsontext.Decoder
	reader        io.ReadCloser
}

func NewJsonStreamReader(separator string, depth int, reader io.ReadCloser) io.ReadCloser {
	return &jsonStreamReader{
		separator: separator,
		depth:     depth,
		decoder:   jsontext.NewDecoder(reader),
		reader:    reader,
	}
}

func (j *jsonStreamReader) Read(p []byte) (n int, err error) {
	n = copy(p, j.buf)
	j.buf = j.buf[n:]

	if n > 0 {
		return
	}

	var str string
	if j.currentLength > 0 {
		str = j.separator
	}

	for {
		token, err := j.decoder.ReadToken()
		if err != nil {
			return 0, err
		}

		switch token.Kind() {
		case jsontext.KindBeginArray, jsontext.KindBeginObject:
			j.currentDepth++
		case jsontext.KindEndArray, jsontext.KindEndObject:
			j.currentDepth--
		}

		if j.currentDepth != j.depth {
			continue
		}

		if token.Kind() == jsontext.KindString {
			str += token.String()
			break
		}
	}

	n += copy(p[n:], str)
	j.buf = []byte(str)[n:]
	j.currentLength++
	return
}

func (j *jsonStreamReader) Close() error {
	return j.reader.Close()
}
