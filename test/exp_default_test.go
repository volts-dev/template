package test

import (
	"bytes"
	//	"bytes"
	"fmt"
	"testing"
	"webgo/template"
)

// ~ website.id if (editable or translatable) and website else None
func TestExpressionDefaultTest(t *testing.T) {
	ctx := make(template.TContext)
	ctx["website"] = map[string]string{"id": "1"}
	ctx["editable"] = true
	ctx["translatable"] = false
	ctx["title"] = "2423"
	exctx := &template.ExecutionContext{
		//		template: nil,

		Public:     ctx,
		Private:    nil,
		Autoescape: true,
	}
	p := template.NewParser(nil)
	node, err := p.ParseDocument(`{% if editable %}{{title}}fasdf{% endif %}`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//vals, err := node.Execute(exctx,)
	//	fmt.Println("222", vals, err)

	//wt := new(templateWriter)
	var str string
	//var b bytes.Buffer
	buf := bytes.NewBuffer(nil)
	err = node.Execute(exctx, buf)
	buf.WriteString(str)
	fmt.Println("ffff", err, str)

}

// ~ website.id if (editable or translatable) and website else None
func TestExpressionDefault2Test(t *testing.T) {
	ctx := make(template.TContext)
	ctx["website"] = map[string]string{"id": "1"}
	ctx["editable"] = true
	ctx["translatable"] = false
	ctx["title"] = "2423"
	//exctx := newExecutionContext(nil, ctx)

	p := template.NewExpressionParser()
	node, err := p.ParseDocument(`{% if (editable or translatable) and website %}{{website.id}}{% endif %}`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	vals, err := node.Evaluate(ctx)
	fmt.Println("222", vals, err)
	/*
		wt := new(templateWriter)
		var str string
		//var b bytes.Buffer
		wt.w = bytes.NewBuffer(nil)
		err = node.Execute(exctx, wt)
		wt.WriteString(str)
		fmt.Println("fff", err.Error(), str)
	*/
}
