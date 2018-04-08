package template

import (
	//	"bytes"
	"fmt"
	"testing"
)

func TestExpressionVarParser(t *testing.T) {
	ctx := make(TContext)
	ctx["a"] = map[string]string{"gg": "bor"}

	p := NewExpressionParser()
	node, err := p.ParseDocument(`{{ a.gg }}`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	vals, err := node.Evaluate(ctx)
	fmt.Println("111", vals, err)
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
func aa(idx, count int, key, value *Value) bool {
	fmt.Println("222", idx, count, key, value)
	return false
}
func TestExpressionConParser(t *testing.T) {
	ctx := make(TContext)
	ctx["website"] = map[string]string{"user_id": "1", "user_id2": "52345234"}
	ctx["user_id"] = "1"
	ctx["user_ids"] = 2
	ctx["title"] = "2423"
	ctx["array"] = []string{"24asd23", "242fasd3", "242asdf3"}
	//exctx := newExecutionContext(nil, ctx)

	p := NewExpressionParser()
	node, err := p.ParseDocument(`{{website.user_id2 if website.user_id == '2' else None }}`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	vals, err := node.Evaluate(ctx)
	fmt.Println("len", len(vals))
	if len(vals) > 0 {
		fmt.Println("val", vals[0])
		fmt.Println("len2222", vals[0].IsNumber(), vals[0].IsString(), vals[0].IsBool(), vals[0].val)
		vals[0].Iterate(func(idx, count int, key, value *Value) bool {
			fmt.Println("222", idx, count, key, value)
			return true
		}, nil)
	}

	fmt.Println("222", err)
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
