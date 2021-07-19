package template

import (
	"bytes"
	"fmt"
	"html"
	"testing"
)

// ~ website.id if (editable or translatable) and website else None
func TestExpressionDefaultTest(t *testing.T) {
	ctx := make(TContext)
	ctx["website"] = map[string]string{"id": "1"}
	ctx["editable"] = true
	ctx["translatable"] = false
	ctx["title"] = "2423"
	ctx["hello"] = "hello"
	ctx["func"] = func() string {
		return "func"
	}

	ctx["lst"] = []string{"123", "12312"}
	//ctx["world"] = "world"
	exctx := &ExecutionContext{
		//		template: nil,

		Public:     ctx,
		Private:    nil,
		Autoescape: true,
	}
	p := NewParser(nil)

	fmt.Println(html.UnescapeString(`{{&apos;123&apos; in lst}}`))
	node, err := p.ParseDocument(html.UnescapeString(`{{&apos;1231&apos; in lst}}`))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//t := new(templateWriter)
	//ar str string
	//var b bytes.Buffer
	buf := bytes.NewBuffer(nil)
	err = node.Execute(exctx, buf)
	fmt.Println("result", buf.String())
	fmt.Println("----------------------")
	vals, err := node.Evaluate(ctx)
	vals[0].Iterate(func(idx, count int, key, value *Value) bool {
		fmt.Println("enum.Iterate for ", idx, count, key, value)
		return true
	}, nil)
	fmt.Println("result", vals, err)
}
