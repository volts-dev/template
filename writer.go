package template

import (
	//"fmt"
	"io"
)

type ITemplateWriter interface {
	io.Writer
	WriteString(string) (int, error)
}

type templateWriter struct {
	w io.Writer
}

func (tw *templateWriter) WriteString(s string) (int, error) {
	return tw.w.Write([]byte(s))
}

func (tw *templateWriter) Write(b []byte) (int, error) {
	return tw.w.Write(b)
}
