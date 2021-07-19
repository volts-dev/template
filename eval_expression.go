package template

import (
	"fmt"
	"math"
	//	"reflect"
)

type (
	// exp root
	TExpression struct {
		// TODO: Add location token?
		expr1    IEvaluator
		expr2    IEvaluator
		operator *TToken
		//tag      string // 表示判断是否返回值或者布尔条件结果
	}

	relationalExpression struct {
		// TODO: Add location token?
		expr1    IEvaluator
		expr2    IEvaluator
		operator *TToken
	}

	simpleExpression struct {
		negate       bool
		negativeSign bool
		term1        IEvaluator
		term2        IEvaluator
		operator     *TToken
	}

	term struct {
		// TODO: Add location token?
		factor1  IEvaluator
		factor2  IEvaluator
		operator *TToken
	}

	power struct {
		// TODO: Add location token?
		power1 IEvaluator
		power2 IEvaluator
	}
)

func (expr *TExpression) FilterApplied(name string) bool {
	return expr.expr1.FilterApplied(name) && (expr.expr2 == nil ||
		(expr.expr2 != nil && expr.expr2.FilterApplied(name)))
}

func (expr *relationalExpression) FilterApplied(name string) bool {
	return expr.expr1.FilterApplied(name) && (expr.expr2 == nil ||
		(expr.expr2 != nil && expr.expr2.FilterApplied(name)))
}

func (expr *simpleExpression) FilterApplied(name string) bool {
	return expr.term1.FilterApplied(name) && (expr.term2 == nil ||
		(expr.term2 != nil && expr.term2.FilterApplied(name)))
}

func (expr *term) FilterApplied(name string) bool {
	return expr.factor1.FilterApplied(name) && (expr.factor2 == nil ||
		(expr.factor2 != nil && expr.factor2.FilterApplied(name)))
}

func (expr *power) FilterApplied(name string) bool {
	return expr.power1.FilterApplied(name) && (expr.power2 == nil ||
		(expr.power2 != nil && expr.power2.FilterApplied(name)))
}

func (expr *TExpression) GetPositionToken() *TToken {
	return expr.expr1.GetPositionToken()
}

func (expr *relationalExpression) GetPositionToken() *TToken {
	return expr.expr1.GetPositionToken()
}

func (expr *simpleExpression) GetPositionToken() *TToken {
	return expr.term1.GetPositionToken()
}

func (expr *term) GetPositionToken() *TToken {
	return expr.factor1.GetPositionToken()
}

func (expr *power) GetPositionToken() *TToken {
	return expr.power1.GetPositionToken()
}

func (expr *TExpression) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	value, err := expr.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

func (expr *relationalExpression) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	value, err := expr.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

func (expr *simpleExpression) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	value, err := expr.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

func (expr *term) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	value, err := expr.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

func (expr *power) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	value, err := expr.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

// # 执行表达式
func (expr *TExpression) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	v1, err := expr.expr1.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	//fmt.Println("TExpression", reflect.TypeOf(expr.expr1), v1.Interface())
	if expr.expr2 != nil {
		v2, err := expr.expr2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		//fmt.Println("TExpression2", reflect.TypeOf(expr.expr2), v2.Interface())
		switch expr.operator.Val {
		case "if":
			if v2.IsTrue() {
				return AsValue(v1.Interface()), nil
			}
			return AsValue(false), nil
		case "else":
			if v1.IsTrue() {
				return AsValue(v2.Interface()), nil
			}
			return AsValue(false), nil

		case "and", "&&":
			//fmt.Println("AND", v1.Interface(), v2.Interface())
			if v1.IsBool() && !v2.IsBool() {
				if v1.IsTrue() {
					return AsValue(v2.Interface()), nil
				}
				return AsValue(v1.IsTrue()), nil
			}

			return AsValue(v1.IsTrue() && v2.IsTrue()), nil

		case "or", "||": // # 新添加以区分返回值还是布尔值
			// # 当都是布尔类型时返回布尔类型
			if (v1.IsBool()) && v2.IsBool() {
				return AsValue(v1.IsTrue() || v2.IsTrue()), nil
			}

			//fmt.Println("OR", v1.IsBool(), v2.IsBool(), expr.operator.Val, v1.Interface(), v2.Interface())
			if v1.IsNil() {
				if v2.IsNil() { //# 当V1和V2都是Nil时返回false
					return AsValue(false), nil
				}
				// # 否则返回V2
				return AsValue(v2.Interface()), nil

			} else {
				if v1.IsBool() && !v1.IsTrue() {
					return AsValue(v2.Interface()), nil
				}

				return AsValue(v1.Interface()), nil
			}
		default:
			panic(fmt.Sprintf("unimplemented: %s", expr.operator.Val))
		}

	}

	return v1, nil
}

func (expr *relationalExpression) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	v1, err := expr.expr1.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if expr.expr2 != nil {
		v2, err := expr.expr2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		switch expr.operator.Val {
		case "<=":
			if v1.IsFloat() || v2.IsFloat() {
				return AsValue(v1.Float() <= v2.Float()), nil
			}
			return AsValue(v1.Integer() <= v2.Integer()), nil
		case ">=":
			if v1.IsFloat() || v2.IsFloat() {
				return AsValue(v1.Float() >= v2.Float()), nil
			}
			return AsValue(v1.Integer() >= v2.Integer()), nil
		case "==":
			return AsValue(v1.EqualValueTo(v2)), nil
		case ">":
			if v1.IsFloat() || v2.IsFloat() {
				return AsValue(v1.Float() > v2.Float()), nil
			}
			return AsValue(v1.Integer() > v2.Integer()), nil
		case "<":
			if v1.IsFloat() || v2.IsFloat() {
				return AsValue(v1.Float() < v2.Float()), nil
			}
			return AsValue(v1.Integer() < v2.Integer()), nil
		case "!=", "<>":
			return AsValue(!v1.EqualValueTo(v2)), nil
		case "in":
			return AsValue(v2.Contains(v1)), nil
		default:
			panic(fmt.Sprintf("unimplemented: %s", expr.operator.Val))
		}
	} else {
		return v1, nil
	}
}

func (expr *simpleExpression) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	t1, err := expr.term1.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	result := t1

	if expr.negate {
		result = result.Negate()
	}

	if expr.negativeSign {
		if result.IsNumber() {
			switch {
			case result.IsFloat():
				result = AsValue(-1 * result.Float())
			case result.IsInteger():
				result = AsValue(-1 * result.Integer())
			default:
				panic("not possible")
			}
		} else {
			return nil, ctx.Error("Negative sign on a non-number expression", expr.GetPositionToken())
		}
	}
	//fmt.Println("simpleExpression", t1.Interface())
	if expr.term2 != nil {
		t2, err := expr.term2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		switch expr.operator.Val {
		case "+":
			if result.IsFloat() || t2.IsFloat() {
				// Result will be a float
				return AsValue(result.Float() + t2.Float()), nil
			}
			// Result will be an integer
			return AsValue(result.Integer() + t2.Integer()), nil
		case "-":
			if result.IsFloat() || t2.IsFloat() {
				// Result will be a float
				return AsValue(result.Float() - t2.Float()), nil
			}
			// Result will be an integer
			return AsValue(result.Integer() - t2.Integer()), nil
		default:
			panic("unimplemented")
		}
	}

	return result, nil
}

func (expr *term) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	f1, err := expr.factor1.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	//fmt.Println("term", f1.Interface())
	if expr.factor2 != nil {
		f2, err := expr.factor2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		switch expr.operator.Val {
		case "*":
			if f1.IsFloat() || f2.IsFloat() {
				// Result will be float
				return AsValue(f1.Float() * f2.Float()), nil
			}
			// Result will be int
			return AsValue(f1.Integer() * f2.Integer()), nil
		case "/":
			if f1.IsFloat() || f2.IsFloat() {
				// Result will be float
				return AsValue(f1.Float() / f2.Float()), nil
			}
			// Result will be int
			return AsValue(f1.Integer() / f2.Integer()), nil
		case "%":
			// Result will be int
			return AsValue(f1.Integer() % f2.Integer()), nil
		default:
			panic("unimplemented")
		}
	} else {
		return f1, nil
	}
}

func (expr *power) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	p1, err := expr.power1.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	//fmt.Println("power", p1.Interface())
	if expr.power2 != nil {
		p2, err := expr.power2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		return AsValue(math.Pow(p1.Float(), p2.Float())), nil
	}
	return p1, nil
}
