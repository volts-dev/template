package template

type (
	INode interface {
		Execute(*ExecutionContext, TemplateWriter) *Error
	}

	INodeTag interface {
		INode
	}

	IEvaluative interface {
		INode
		Evaluate(*ExecutionContext) (*Value, *Error)
	}

	IEvaluator interface {
		INode
		GetPositionToken() *TToken
		Evaluate(*ExecutionContext) (*Value, *Error)
		FilterApplied(name string) bool
	}

	// The root document
	nodeDocument struct {
		Nodes []INode
	}

	nodeHTML struct {
		token *TToken
	}

	NodeWrapper struct {
		Endtag string
		nodes  []INode
	}

	nodeVariable struct {
		locationToken *TToken
		expr          IEvaluator
	}
)

func (doc *nodeDocument) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	for _, n := range doc.Nodes {
		err := n.Execute(ctx, writer)
		if err != nil {
			return err
		}
	}
	return nil
}

// #执行表达式
func (doc *nodeDocument) Evaluate(context TContext) (res_val []*Value, res_err *Error) {
	//fmt.Println("Evaluate:1", len(doc.Nodes))
	ctx := newExecutionContext(nil, context)
	res_val = make([]*Value, 0)
	var val *Value
	for _, n := range doc.Nodes {
		//fmt.Println("Evaluate2:", n, reflect.TypeOf(n))
		/*
			if eval, allowed := n.(*nodeVariable); allowed {
				val, err = eval.Evaluate(ctx)
			} else if eval, allowed := n.(IEvaluative); allowed {
				val, err = eval.Evaluate(ctx)
			}*/
		if eval, allowed := n.(IEvaluative); allowed {
			val, res_err = eval.Evaluate(ctx)
			if res_err != nil {
				//fmt.Println("fasdfasd", res_err.Error())
				return
			}

			res_val = append(res_val, val)
		}
		//fmt.Println("Evaluate23", val, res_err == nil, res_err)
	}

	return res_val, nil
}

func (n *nodeHTML) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	writer.WriteString(n.token.Val)
	return nil
}

func (wrapper *NodeWrapper) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	for _, n := range wrapper.nodes {
		err := n.Execute(ctx, writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (nv *nodeVariable) FilterApplied(name string) bool {
	return nv.expr.FilterApplied(name)
}

func (nv *nodeVariable) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	return nv.expr.Evaluate(ctx)
}

func (nv *nodeVariable) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	value, err := nv.expr.Evaluate(ctx)
	if err != nil {
		return err
	}

	if !nv.expr.FilterApplied("safe") && !value.safe && value.IsString() && ctx.Autoescape {
		// apply escape filter
		value, err = filters["escape"](value, nil)
		if err != nil {
			return err
		}
	}

	writer.WriteString(value.String())
	return nil
}
