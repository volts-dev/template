package template

import (
	"fmt"
	"io"
)

type TemplateWriter interface {
	io.Writer
	WriteString(string) (int, error)
}

type templateWriter struct {
	w io.Writer
}

func (tw *templateWriter) WriteString(s string) (int, error) {
	fmt.Println("77", tw, s)
	fmt.Println("777", tw, tw.w)
	return tw.w.Write([]byte(s))
}

func (tw *templateWriter) Write(b []byte) (int, error) {
	return tw.w.Write(b)
}
