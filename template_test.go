package template

import (
	"bytes"
	"fmt"
	"testing"
)

// ~ website.id if (editable or translatable) and website else None
func TestExpressionDefaultTest(t *testing.T) {
	ctx := make(TContext)
	ctx["website"] = map[string]string{"id": "1"}
	ctx["editable"] = true
	ctx["translatable"] = false
	ctx["title"] = "2423"
	exctx := &ExecutionContext{
		//		template: nil,

		Public:     ctx,
		Private:    nil,
		Autoescape: true,
	}
	p := NewParser(nil)
	node, err := p.ParseDocument(`{% if editable %}{{title}}fasdf{% endif %}`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//t := new(templateWriter)
	//ar str string
	//var b bytes.Buffer
	buf := bytes.NewBuffer(nil)
	err = node.Execute(exctx, buf)
	fmt.Println("fffa", buf.String())
	//t.WriteString(str)
	//mt.Println("fff", err)

}
