package template

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	varTypeInt = iota
	varTypeIdent
)

type variablePart struct {
	typ int
	s   string
	i   int

	isFunctionCall bool
	callingArgs    []functionCallArgument // needed for a function call, represents all argument nodes (INode supports nested function calls)
}

type functionCallArgument interface {
	Evaluate(*ExecutionContext) (*Value, *Error)
}

// TODO: Add location tokens
type stringResolver struct {
	locationToken *TToken
	val           string
}

type intResolver struct {
	locationToken *TToken
	val           int
}

type floatResolver struct {
	locationToken *TToken
	val           float64
}

type boolResolver struct {
	locationToken *TToken
	val           bool
}

type variableResolver struct {
	locationToken *TToken

	parts []*variablePart
}

type nodeFilteredVariable struct {
	locationToken *TToken

	resolver    IEvaluator
	filterChain []*filterCall
}

type nodeIf struct {
	locationToken *TToken
	conditions    []IEvaluator
	wrappers      []IEvaluative
}

func (v *nodeFilteredVariable) Execute(ctx *ExecutionContext, writer ITemplateWriter) *Error {
	value, err := v.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

func (vr *variableResolver) Execute(ctx *ExecutionContext, writer ITemplateWriter) *Error {
	value, err := vr.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

func (s *stringResolver) Execute(ctx *ExecutionContext, writer ITemplateWriter) *Error {
	value, err := s.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

func (i *intResolver) Execute(ctx *ExecutionContext, writer ITemplateWriter) *Error {
	value, err := i.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

func (f *floatResolver) Execute(ctx *ExecutionContext, writer ITemplateWriter) *Error {
	value, err := f.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

func (b *boolResolver) Execute(ctx *ExecutionContext, writer ITemplateWriter) *Error {
	value, err := b.Evaluate(ctx)
	if err != nil {
		return err
	}
	writer.WriteString(value.String())
	return nil
}

func (v *nodeFilteredVariable) GetPositionToken() *TToken {
	return v.locationToken
}

func (vr *variableResolver) GetPositionToken() *TToken {
	return vr.locationToken
}

func (s *stringResolver) GetPositionToken() *TToken {
	return s.locationToken
}

func (i *intResolver) GetPositionToken() *TToken {
	return i.locationToken
}

func (f *floatResolver) GetPositionToken() *TToken {
	return f.locationToken
}

func (b *boolResolver) GetPositionToken() *TToken {
	return b.locationToken
}

func (s *stringResolver) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	return AsValue(s.val), nil
}

func (i *intResolver) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	return AsValue(i.val), nil
}

func (f *floatResolver) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	return AsValue(f.val), nil
}

func (b *boolResolver) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	return AsValue(b.val), nil
}

func (s *stringResolver) FilterApplied(name string) bool {
	return false
}

func (i *intResolver) FilterApplied(name string) bool {
	return false
}

func (f *floatResolver) FilterApplied(name string) bool {
	return false
}

func (b *boolResolver) FilterApplied(name string) bool {
	return false
}

func (vr *variableResolver) FilterApplied(name string) bool {
	return false
}

func (vr *variableResolver) String() string {
	parts := make([]string, 0, len(vr.parts))
	for _, p := range vr.parts {
		switch p.typ {
		case varTypeInt:
			parts = append(parts, strconv.Itoa(p.i))
		case varTypeIdent:
			parts = append(parts, p.s)
		default:
			panic("unimplemented")
		}
	}
	return strings.Join(parts, ".")
}

func (vr *variableResolver) resolve(ctx *ExecutionContext) (*Value, error) {
	var current reflect.Value
	var isSafe bool

	for idx, part := range vr.parts {
		if idx == 0 {
			// We're looking up the first part of the variable.
			// First we're having a look in our private
			// context (e. g. information provided by tags, like the forloop)
			val, inPrivate := ctx.Private[vr.parts[0].s]
			if !inPrivate {
				// Nothing found? Then have a final lookup in the public context
				val = ctx.Public[vr.parts[0].s]
			}
			current = reflect.ValueOf(val) // Get the initial value
		} else {
			// Next parts, resolve it from current

			// Before resolving the pointer, let's see if we have a method to call
			// Problem with resolving the pointer is we're changing the receiver
			isFunc := false
			if part.typ == varTypeIdent {
				funcValue := current.MethodByName(part.s)
				if funcValue.IsValid() {
					current = funcValue
					isFunc = true
				}
			}

			if !isFunc {
				// If current a pointer, resolve it
				if current.Kind() == reflect.Ptr {
					current = current.Elem()
					if !current.IsValid() {
						// Value is not valid (anymore)
						return AsValue(nil), nil
					}
				}

				// Look up which part must be called now
				switch part.typ {
				case varTypeInt:
					// Calling an index is only possible for:
					// * slices/arrays/strings
					switch current.Kind() {
					case reflect.String, reflect.Array, reflect.Slice:
						if part.i >= 0 && current.Len() > part.i {
							current = current.Index(part.i)
						} else {
							// Access to non-existed index will cause an error
							return nil, fmt.Errorf("%s is out of range (type: %s, length: %d)",
								vr.String(), current.Kind().String(), current.Len())
						}
					default:
						return nil, fmt.Errorf("Can't access an index on type %s (variable %s)",
							current.Kind().String(), vr.String())
					}
				case varTypeIdent:
					// debugging:
					// fmt.Printf("now = %s (kind: %s)\n", part.s, current.Kind().String())

					// Calling a field or key
					switch current.Kind() {
					case reflect.Struct:
						current = current.FieldByName(part.s)
					case reflect.Map:
						current = current.MapIndex(reflect.ValueOf(part.s))
					default:
						return nil, fmt.Errorf("Can't access a field by name on type %s (variable %s)",
							current.Kind().String(), vr.String())
					}
				default:
					panic("unimplemented")
				}
			}
		}

		if !current.IsValid() {
			// Value is not valid (anymore)
			return AsValue(nil), nil
		}

		// If current is a reflect.ValueOf(pongo2.Value), then unpack it
		// Happens in function calls (as a return value) or by injecting
		// into the execution context (e.g. in a for-loop)
		if current.Type() == reflect.TypeOf(&Value{}) {
			tmpValue := current.Interface().(*Value)
			current = tmpValue.val
			isSafe = tmpValue.safe
		}

		// Check whether this is an interface and resolve it where required
		if current.Kind() == reflect.Interface {
			current = reflect.ValueOf(current.Interface())
		}

		// Check if the part is a function call
		if part.isFunctionCall || current.Kind() == reflect.Func {
			// Check for callable
			if current.Kind() != reflect.Func {
				return nil, fmt.Errorf("'%s' is not a function (it is %s)", vr.String(), current.Kind().String())
			}

			// Check for correct function syntax and types
			// func(*Value, ...) *Value
			t := current.Type()

			// Input arguments
			if len(part.callingArgs) != t.NumIn() && !(len(part.callingArgs) >= t.NumIn()-1 && t.IsVariadic()) {
				return nil,
					fmt.Errorf("Function input argument count (%d) of '%s' must be equal to the calling argument count (%d).",
						t.NumIn(), vr.String(), len(part.callingArgs))
			}

			// Output arguments
			if t.NumOut() != 1 {
				return nil, fmt.Errorf("'%s' must have exactly 1 output argument", vr.String())
			}

			// Evaluate all parameters
			var parameters []reflect.Value

			numArgs := t.NumIn()
			isVariadic := t.IsVariadic()
			var fnArg reflect.Type

			for idx, arg := range part.callingArgs {
				pv, err := arg.Evaluate(ctx)
				if err != nil {
					return nil, err
				}

				if isVariadic {
					if idx >= t.NumIn()-1 {
						fnArg = t.In(numArgs - 1).Elem()
					} else {
						fnArg = t.In(idx)
					}
				} else {
					fnArg = t.In(idx)
				}

				if fnArg != reflect.TypeOf(new(Value)) {
					// Function's argument is not a *pongo2.Value, then we have to check whether input argument is of the same type as the function's argument
					if !isVariadic {
						if fnArg != reflect.TypeOf(pv.Interface()) && fnArg.Kind() != reflect.Interface {
							return nil, fmt.Errorf("Function input argument %d of '%s' must be of type %s or *pongo2.Value (not %T).",
								idx, vr.String(), fnArg.String(), pv.Interface())
						}
						// Function's argument has another type, using the interface-value
						parameters = append(parameters, reflect.ValueOf(pv.Interface()))
					} else {
						if fnArg != reflect.TypeOf(pv.Interface()) && fnArg.Kind() != reflect.Interface {
							return nil, fmt.Errorf("Function variadic input argument of '%s' must be of type %s or *pongo2.Value (not %T).",
								vr.String(), fnArg.String(), pv.Interface())
						}
						// Function's argument has another type, using the interface-value
						parameters = append(parameters, reflect.ValueOf(pv.Interface()))
					}
				} else {
					// Function's argument is a *pongo2.Value
					parameters = append(parameters, reflect.ValueOf(pv))
				}
			}

			// Check if any of the values are invalid
			for _, p := range parameters {
				if p.Kind() == reflect.Invalid {
					return nil, fmt.Errorf("Calling a function using an invalid parameter")
				}
			}

			// Call it and get first return parameter back
			rv := current.Call(parameters)[0]

			if rv.Type() != reflect.TypeOf(new(Value)) {
				current = reflect.ValueOf(rv.Interface())
			} else {
				// Return the function call value
				current = rv.Interface().(*Value).val
				isSafe = rv.Interface().(*Value).safe
			}
		}

		if !current.IsValid() {
			// Value is not valid (e. g. NIL value)
			return AsValue(nil), nil
		}
	}

	return &Value{val: current, safe: isSafe}, nil
}

func (vr *variableResolver) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	value, err := vr.resolve(ctx)
	if err != nil {
		return AsValue(nil), ctx.Error(err.Error(), vr.locationToken)
	}
	return value, nil
}

func (v *nodeFilteredVariable) FilterApplied(name string) bool {
	for _, filter := range v.filterChain {
		if filter.name == name {
			return true
		}
	}
	return false
}

func (v *nodeFilteredVariable) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	value, err := v.resolver.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	for _, filter := range v.filterChain {
		value, err = filter.Execute(value, ctx)
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

// var if expr else not
func (node *nodeIf) Evaluate(ctx *ExecutionContext) (*Value, *Error) {
	fmt.Println("Evaluate node1", len(node.conditions), len(node.wrappers))
	for i, condition := range node.conditions {
		//fmt.Println("Evaluate node", condition)
		result, err := condition.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		//fmt.Println("Evaluate node3", result)
		if result.IsTrue() {
			return node.wrappers[i].Evaluate(ctx)
		}

		// Last value
		if len(node.wrappers) > i+1 {
			//fmt.Println("Evaluate node4", node.wrappers[i+1])
			return node.wrappers[i+1].Evaluate(ctx)
		} else {
			return AsValue(nil), nil
		}

	}

	return nil, nil
}

func (node *nodeIf) Execute(ctx *ExecutionContext, writer ITemplateWriter) *Error {
	for i, condition := range node.conditions {
		result, err := condition.Evaluate(ctx)
		if err != nil {
			return err
		}

		if result.IsTrue() {
			return node.wrappers[i].Execute(ctx, writer)
		}

		// Last value
		if len(node.wrappers) > i+1 {
			return node.wrappers[i+1].Execute(ctx, writer)
		}

	}

	return nil
}

func (s *nodeIf) FilterApplied(name string) bool {
	return false
}

func (v *nodeIf) GetPositionToken() *TToken {
	return v.locationToken
}
