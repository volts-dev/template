package template

import (
	"fmt"
	"math"
)

/*

#模板全局默认变量
	TExpression
	│
	├─module 应用模块目录
	│  ├─web 模块目录
	│  │  ├─staric 静态资源目录
	│  │  │   ├─uploads 上传根目录
	│  │  │   ├─lib 资源库文件目录(常用作前端框架库)
	│  │  │   └─src 资源文件
	│  │  │      ├─js 资源Js文件目录
	│  │  │      ├─img 资源图片文件目录
	│  │  │      └─css 资源Css文件
	│  │  ├─model 模型目录
	│  │  ├─template 视图文件目录
	│  │  ├─data 数据目录
	│  │  ├─model 模型目录
	│  │  └─controller.go 控制器
	│  │
	│  ├─base 模块目录
	│  │
	│  └─... 扩展的可装卸功能模块或插件
	│
	├─staric 静态资源目录
	│  ├─uploads 上传根目录
	│  ├─lib 资源库文件目录(常用作前端框架库)
	│  └─src 资源文件
	│     ├─js 资源Js文件目录
	│     ├─img 资源图片文件目录
	│     └─css 资源Css文件
	│
	├─deploy 部署文件目录
	│
	├─main.go 主文件
	└─main.ini 配置文件
*/

type (
	// exp root
	TExpression struct {
		// TODO: Add location token?
		expr1    IEvaluator
		expr2    IEvaluator
		operator *TToken
		tag      string // 表示判断是否返回值或者布尔条件结果
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

func (expr *TExpression) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
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
		case "and", "&&":
			return AsValue(v1.IsTrue() && v2.IsTrue()), nil
		case "or", "||":
			if expr.tag == "{{" { // 新添加以区分返回值还是布尔值
				if !v1.IsNil() {
					return AsValue(v1.Interface()), nil
				} else {
					return AsValue(v2.Interface()), nil
				}
			} else {
				return AsValue(v1.IsTrue() || v2.IsTrue()), nil
			}

		default:
			panic(fmt.Sprintf("unimplemented: %s", expr.operator.Val))
		}
	} else {
		return v1, nil
	}
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
	if expr.power2 != nil {
		p2, err := expr.power2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		return AsValue(math.Pow(p1.Float(), p2.Float())), nil
	}
	return p1, nil
}
