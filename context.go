package template

import (
	"bytes"
	"fmt"
	"regexp"
)

var reIdentifiers = regexp.MustCompile("^[a-zA-Z0-9_]+$")

// A Context type provides constants, variables, instances or functions to a template.
//
// pongo2 automatically provides meta-information or functions through the "pongo2"-key.
// Currently, context["pongo2"] contains the following keys:
//  1. version: returns the version string
//
// Template examples for accessing items from your context:
//     {{ myconstant }}
//     {{ myfunc("test", 42) }}
//     {{ user.name }}
//     {{ pongo2.version }}
type TContext map[string]interface{}

func NewContext() TContext {
	return make(TContext)
}
func (self TContext) checkForValidIdentifiers() *Error {
	for k, v := range self {
		if !reIdentifiers.MatchString(k) {
			return &Error{
				Sender:   "checkForValidIdentifiers",
				ErrorMsg: fmt.Sprintf("Context-key '%s' (value: '%+v') is not a valid identifier.", k, v),
			}
		}
	}
	return nil
}

// Update updates this context with the key/value-pairs from another context.
func (self TContext) Update(other TContext) TContext {
	for k, v := range other {
		self[k] = v
	}
	return self
}

func (self TContext) Copy() TContext {
	ctx := TContext{}

	for key, val := range self {
		ctx[key] = val
	}

	return ctx
}

// #
// @mode: expr/con 简单表达式或者条件表达式

func (self TContext) SafeEval(mode string, expr string) ([]*Value, error) {
	/*
	   locals_dict = collections.defaultdict(lambda: None)
	   locals_dict.update(self)
	   locals_dict.pop('cr', None)
	   locals_dict.pop('loader', None)
	   return eval(expr, None, locals_dict, nocopy=True, locals_builtins=True)*/

	p := NewExpressionParser()
	node, err := p.ParseDocument(fmt.Sprintf(`{{ %s }}`, expr))
	if err != nil {
		//fmt.Println("fasd", err.Error())
		return nil, err
	}

	if mode == "con" {
		exctx := &ExecutionContext{
			Public:     self,
			Private:    nil,
			Autoescape: true,
		}

		buf := bytes.NewBuffer(nil)
		err := node.Execute(exctx, buf)
		if err != nil {
			return nil, err
		}

		vals := make([]*Value, 0)
		vals = append(vals, AsValue(buf.String()))
		return vals, nil

	} else {

		vals, err := node.Evaluate(self)
		if err != nil {
			return nil, err
		}
		return vals, nil

	}

	return nil, nil
}

// ExecutionContext contains all data important for the current rendering state.
//
// If you're writing a custom tag, your tag's Execute()-function will
// have access to the ExecutionContext. This struct stores anything
// about the current rendering process's Context including
// the Context provided by the user (field Public).
// You can safely use the Private context to provide data to the user's
// template (like a 'forloop'-information). The Shared-context is used
// to share data between tags. All ExecutionContexts share this context.
//
// Please be careful when accessing the Public data.
// PLEASE DO NOT MODIFY THE PUBLIC CONTEXT (read-only).
//
// To create your own execution context within tags, use the
// NewChildExecutionContext(parent) function.
type ExecutionContext struct {
	template *TTemplate

	Autoescape bool
	Public     TContext
	Private    TContext
	Shared     TContext
}

var pongo2MetaContext = TContext{
	"version": Version,
}

func newExecutionContext(tpl *TTemplate, ctx TContext) *ExecutionContext {
	privateCtx := make(TContext)

	// Make the pongo2-related funcs/vars available to the context
	privateCtx["pongo2"] = pongo2MetaContext

	return &ExecutionContext{
		template: tpl,

		Public:     ctx,
		Private:    privateCtx,
		Autoescape: true,
	}
}

func NewChildExecutionContext(parent *ExecutionContext) *ExecutionContext {
	newctx := &ExecutionContext{
		template: parent.template,

		Public:     parent.Public,
		Private:    make(TContext),
		Autoescape: parent.Autoescape,
	}
	newctx.Shared = parent.Shared

	// Copy all existing private items
	newctx.Private.Update(parent.Private)

	return newctx
}

func (ctx *ExecutionContext) Error(msg string, token *TToken) *Error {
	filename := ctx.template.Name
	var line, col int
	if token != nil {
		// No tokens available
		// TODO: Add location (from where?)
		filename = token.Filename
		line = token.Line
		col = token.Col
	}
	return &Error{
		Template: ctx.template,
		Filename: filename,
		Line:     line,
		Column:   col,
		Token:    token,
		Sender:   "execution",
		ErrorMsg: msg,
	}
}

func (ctx *ExecutionContext) Logf(format string, args ...interface{}) {
	//	ctx.template.set.logf(format, args...)
}
