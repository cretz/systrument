package util

import (
	"bytes"
	"fmt"
	"io"
	"text/scanner"
)

func NewCommentStrippedJSONReader(r io.Reader) io.Reader {
	// I think we can just use the scanner here
	ret := &commentStrippedJSONReader{
		scan: &scanner.Scanner{},
		buf:  &bytes.Buffer{},
	}
	ret.scan.Init(r)
	ret.scan.Whitespace = 0
	ret.scan.Error = func(_ *scanner.Scanner, msg string) {
		ret.err = fmt.Errorf("Error reading: %v", msg)
	}
	return ret
}

type commentStrippedJSONReader struct {
	scan *scanner.Scanner
	err  error
	buf  *bytes.Buffer
}

func (r *commentStrippedJSONReader) Read(p []byte) (int, error) {
	for r.buf.Len() < len(p) && r.err == nil {
		if r.scan.Scan() == scanner.EOF {
			r.err = io.EOF
		} else {
			r.buf.WriteString(r.scan.TokenText())
		}
	}
	ret, _ := r.buf.Read(p)
	return ret, r.err
}
