package template

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

type TBlockElement struct {
	name      string
	inverted  bool
	startline int
	elements  []interface{}
}

type TTextElement struct {
	text []byte
}

type TVarElement struct {
	name string
	raw  bool
}

type TSectionElement struct {
	name      string
	inverted  bool
	startline int
	elems     []interface{}
}

type TParserError struct {
	line    int
	message string
}

// Parser 提供解析的逻辑 和数据的变换
type Parser struct {
	name     string
	template *TTemplate
	parent   *Parser
	lex      *lexer
	root     *nodeDocument // 解析后的Node
	silo     []interface{}

	idx       int
	tokens    []*TToken
	lastToken *TToken
}

// TODO 支持<a/>格式
// Creates a new parser to parse tokens.
// Used inside pongo2 to parse documents and to provide an easy-to-use
// parser for tag authors
func newParser(name string, tokens []*TToken, template *TTemplate) *Parser {
	p := &Parser{
		name:     name,
		tokens:   tokens,
		template: template,
	}
	if len(tokens) > 0 {
		p.lastToken = tokens[len(tokens)-1]
	}
	return p
}

// #新建解析器可以独立于TTemplate使用
func NewParser(aTemplete *TTemplate) *Parser {
	parse := &Parser{
		template: aTemplete,
		parent:   nil,
		//		lex:      lLexer,
		silo: []interface{}{}}

	return parse
}

// #废弃
func NewTemplateParser(aTemplete *TTemplate) *Parser {
	lLexer := &lexer{
		//stream:     htmlstream,
		leftdelim:  LEFT_DELIM,
		rightdelim: RIGHT_DELIM,
		pos:        0,
		curline:    1}

	parse := &Parser{
		template: aTemplete,
		parent:   nil,
		lex:      lLexer,
		silo:     []interface{}{}}

	return parse
}

// #废弃
func NewExpressionParser() (res_parser *Parser) {
	/*lLexer := &lexer{
	//stream:     htmlstream,
	input:      expr,
	leftdelim:  LEFT_DELIM,
	rightdelim: RIGHT_DELIM,
	pos:        0,
	curline:    1}
	*/

	res_parser = &Parser{
		template: nil,
		parent:   nil,
		//		lex:      lLexer,
		silo: []interface{}{}}

	//res_parser.tokens, res_err = NewLexer("name", expr) // 解析为
	//res_pareser.parseDocument()                          // 解析为Node
	//	res_pareser.root.Execute()
	//fmt.Println("NewExpressionParser", res_parser.tokens)
	return
}

func (self TParserError) Error() string {
	return fmt.Sprintf("line %d: %s", self.line, self.message)
}

func (self *Parser) NewParse(htmlstream string) (*Parser, error) {
	xlex := &lexer{
		input:      htmlstream,
		leftdelim:  LEFT_DELIM,
		rightdelim: RIGHT_DELIM,
		pos:        0,
		curline:    1}

	parse := &Parser{
		template: self.template,
		parent:   nil,
		lex:      xlex,
		silo:     []interface{}{}}
	return parse, errors.New("NewParse")
}

// s Html文件流
func (self *Parser) parse_block(block *TBlockElement) error {
	for { //#1 循环匹配 符号   "gsdfgsdfgsdf{{"
		text, err := self.lex.scan_string(LEFT_DELIM) //匹配返回{{
		if err == io.EOF {
			//block 里不需要保存非Block 里的数据
			//self.silo = append(self.silo, &TTextElement{[]byte(text)}) // 保存最后本次查询到的字符
			return nil
		}

		//#2 保存Block前段内容到Silo里 保存前剔除字符 "{{"
		text = text[0 : len(text)-len(LEFT_DELIM)]
		block.elements = append(block.elements, &TTextElement{[]byte(text)}) // 再保存"{{"之前的数据

		//#3 以下判断是否为 "{{{" 如果是则匹配到 "}}}"
		if self.lex.pos < len(self.lex.input) && self.lex.input[self.lex.pos] == '{' { //如果POS在文件内且 当前Pos还是={ 说明是{{{模式}}}
			text, err = self.lex.scan_string("}" + RIGHT_DELIM) //查找}}}
		} else {
			text, err = self.lex.scan_string(RIGHT_DELIM) //查找}}
		}
		if err == io.EOF {
			return TParserError{self.lex.curline, "unmatched open tag"}
		}

		//  删除 }} 例如 "  Var }}"//再去空白/如果什么都没有返回
		tag := strings.TrimSpace(text[0 : len(text)-len(RIGHT_DELIM)])
		if len(tag) == 0 {
			return TParserError{self.lex.curline, "empty tag"}
		}
		//tag = strings.ToLower(tag)
		//============ 以上是寻找和分析 Tag 的内容.===============

		tags := strings.Split(tag, " ") // 对Tag进行分割
		//log.Print("[parse_block()]:Split", tags) //################
		switch strings.ToLower(tags[0]) {
		case "trans":
			if len(tags) > 1 {
				str := strings.Replace(tag, "trans", "", -1)
				OrgStr := bytes.Trim([]byte(str), "'\"` ") // #注意必须包括[空格]
				//log.Println(">>>>>>>>>>>>>>>>>>>:", str, string(OrgStr))
				NewStr := self.template.langMap[string(OrgStr)] // 获得翻译后的文本
				if NewStr == "" {
					block.elements = append(block.elements, &TTextElement{OrgStr}) // 没翻译则使用原文
				} else {
					block.elements = append(block.elements, &TTextElement{[]byte(NewStr)})
				}
			}
		case "include":
			if len(tags) > 1 {
				tag_file := strings.Trim(string(tags[1]), "'\"`")
				//fmt.Println("include:", tag_file) //####

				// 获得 {* extends "base.html" *} 文件名
				stream, err := ioutil.ReadFile(filepath.Join(self.template.DirPath, tag_file)) // 读取文件
				if err != nil {
					return err
				}

				// 创建父系Html 的parse
				NewBlock, _ := self.NewParse(string(stream))
				err = NewBlock.parse()
				if err != nil {
					return err
				}

				//分别插入分析好Include文档的元素
				for _, element := range NewBlock.silo {
					block.elements = append(block.elements, element)
					/*
						switch elem := element.(type) {
						case *TBlockElement:
							fmt.Println("NewBlock.silo:TBlockElement", elem.name) //####
						case *TTextElement:
							fmt.Println("NewBlock.silo:TTextElement", string(elem.text)) //####
						}
					*/
				}
			}
		case "extends":
			break //Block 里不会有这个
		case "block":
			if len(tags) == 2 {
				block_name := strings.Trim(string(tags[1]), "'\"`") //获得名字
				//log.Print("[parse_block()]:block/:", block_name)    //###########
				be := TBlockElement{block_name, false, self.lex.curline, []interface{}{}}
				err := self.parse_block(&be)
				if err != nil {
					return err
				}
				block.elements = append(block.elements, &be)
				//log.Print("[parse_block()]:block/", block_name+">>", be) //################

			}
			break
		case "endblock":
			//log.Print("[parse_block()]:endblock/:")
			return nil //******* 这里跳出循环相当重要 不可以使用Break继续For

		default:
			break
		}
	}
}

// s Html文件流
func (self *Parser) parse() error {

	for { //循环匹配 符号   "gsdfgsdfgsdf{{"
		text, err := self.lex.scan_string(LEFT_DELIM) //匹配返回{{
		if err == io.EOF {
			//put the remaining text in a block
			//log.Print("文件结束")

			self.silo = append(self.silo, &TTextElement{[]byte(text)}) // 保存最后本次查询到的字符
			//log.Print("self.parent.silo:", text)

			// 到文文件分析完毕如果是不是Base.Html 则启动模板合并
			if self.parent != nil {
				self.combine_block()
				//log.Print("combine.silo:", self.silo)
				self.silo = self.parent.silo
				//log.Print("combine.parent.silo:", self.parent.silo)
			}
			return nil
		}

		// 保存HTML 到Silo里
		text = text[0 : len(text)-len(LEFT_DELIM)] //剔除字符 "{{"

		self.silo = append(self.silo, &TTextElement{[]byte(text)}) // 再保存"{{"之前的数据

		// 以下判断是否为{{{如果是则匹配到}}}
		if self.lex.pos < len(self.lex.input) && self.lex.input[self.lex.pos] == '{' { //如果POS在文件内且 当前Pos还是={ 说明是{{{模式}}}
			text, err = self.lex.scan_string("}" + RIGHT_DELIM) //查找}}}
		} else {
			text, err = self.lex.scan_string(RIGHT_DELIM) //查找}}
		}

		if err == io.EOF {
			//put the remaining text in a block
			return TParserError{self.lex.curline, "unmatched open tag"}
		}

		//  删除 }} 例如 "  Var }}"/再去空白/如果什么都没有返回
		tag := strings.TrimSpace(text[0 : len(text)-len(RIGHT_DELIM)])
		if len(tag) == 0 {
			return TParserError{self.lex.curline, "empty tag"}
		}
		//tag = strings.ToLower(tag)

		tags := strings.Split(tag, " ")
		//log.Print("[parse()]:Split", tags) //########
		switch strings.ToLower(tags[0]) {
		case "trans":
			if len(tags) > 1 {
				str := strings.Replace(tag, "trans", "", -1)
				OrgStr := bytes.Trim([]byte(str), "'\"` ")      // #注意必须包括[空格]
				NewStr := self.template.langMap[string(OrgStr)] // 获得翻译后的文本
				if NewStr == "" {
					self.silo = append(self.silo, &TTextElement{OrgStr}) // 没翻译则使用原文
				} else {
					self.silo = append(self.silo, &TTextElement{[]byte(NewStr)})
				}
			}
		case "include":
			if len(tags) > 1 {

				tag_file := strings.Trim(string(tags[1]), "'\"`")
				//fmt.Println("include:", tag_file) //####
				// 获得 {* extends "base.html" *} 文件名
				filepath := filepath.Join(self.template.DirPath, tag_file)
				stream, err := ioutil.ReadFile(filepath) // 读取文件
				if err != nil {
					return err
				}

				// 创建父系Html 的parse
				NewBlock, _ := self.NewParse(string(stream))
				err = NewBlock.parse()
				if err != nil {
					return err
				}

				//分别插入分析好Include文档的元素
				for _, element := range NewBlock.silo {
					self.silo = append(self.silo, element)
					/*
						switch elem := element.(type) {
						case *TBlockElement:
							fmt.Println("NewBlock.silo:TBlockElement", elem.name) //####
						case *TTextElement:
							fmt.Println("NewBlock.silo:TTextElement", string(elem.text)) //####
						}
					*/
				}
			}
		case "extends":
			//strings.HasPrefix(string(tags[0]), "extends")
			if len(tags) > 1 {

				tag_file := strings.Trim(string(tags[1]), "'\"`") // 获得 {* extends "base.html" *} 文件名

				filepath := filepath.Join(self.template.DirPath, tag_file)
				//log.Print("extends:", string(tags[1])) //####
				stream, err := ioutil.ReadFile(filepath) // 读取文件
				//log.Print("len(tags)", len(tags), tag_file, string(tags[0]))
				if err != nil {
					return err
				}

				self.silo = append(self.silo, &TTextElement{[]byte(tag_file)})

				// 创建父系Html 的parse
				self.parent, _ = self.NewParse(string(stream))
				err = self.parent.parse()

				if err != nil {
					return err
				}
				//log.Print("[parse()]:extends/", tag_file+">>", self.parent)
			}
		case "block":
			if len(tags) == 2 {
				block_name := strings.Trim(string(tags[1]), "'\"`") //获得名字
				//log.Print("[parse()]:block/:", block_name)          //###########

				be := TBlockElement{block_name, false, self.lex.curline, []interface{}{}}
				err := self.parse_block(&be) //进入树 分析
				if err != nil {
					return err
				}

				self.silo = append(self.silo, &be)
				//log.Print("[parse()]:block/", block_name+">>", be) //###########
			}
			break
		case "endblock":
			//log.Print("[parse()]:endblock/:")

			break

		default:
			break
			/*
			   // 匹配 {{! var}}
			   switch tags[0] {
			   case '!':
			       //ignore comment
			       break
			   case '#', '^': 	//swicth

			       name := strings.TrimSpace(tag[1:]) // tag 名除符号外

			       if len(tmpl.data) > tmpl.p && tmpl.data[tmpl.p] == '\n' { //在范围内 和 }}后面等于\n
			           tmpl.p += 1
			           //在范围内 和
			       } else if len(tmpl.data) > tmpl.p+1 && tmpl.data[tmpl.p] == '\r' && tmpl.data[tmpl.p+1] == '\n' {
			           tmpl.p += 2
			       }

			       se := sectionElement{name, tag[0] == '^', tmpl.curline, []interface{}{}}
			       err := tmpl.parseSection(&se)
			       if err != nil {
			           return err
			       }
			       tmpl.elems = append(tmpl.elems, &se)
			   case '/':
			       return parseError{tmpl.curline, "unmatched close tag"}
			   case '>':
			       name := strings.TrimSpace(tag[1:])
			       partial, err := tmpl.parsePartial(name)
			       if err != nil {
			           return err
			       }
			       tmpl.elems = append(tmpl.elems, partial)
			   case '=':
			       if tag[len(tag)-1] != '=' {
			           return parseError{tmpl.curline, "Invalid meta tag"}
			       }
			       tag = strings.TrimSpace(tag[1 : len(tag)-1])
			       newtags := strings.SplitN(tag, " ", 2)
			       if len(newtags) == 2 {
			           tmpl.otag = newtags[0]
			           tmpl.ctag = newtags[1]
			       }
			   case '{':
			       //use a raw tag
			       if tag[len(tag)-1] == '}' {
			           tmpl.elems = append(tmpl.elems, &varElement{tag[1 : len(tag)-1], true})
			       }
			   default:
			       tmpl.elems = append(tmpl.elems, &varElement{tag, false})
			   } //2 switch
			*/
		} //1 switch
	}

	return nil
}

/*
	func (self *Parser) Next() {
		name, block := self.l.scan()
		self.silo[name] = block
	}


	// parse until one of the following tags
	func (self *Parser) ParseUntil(tags ...string) (string, NodeList) {
		r := make(NodeList, 0, 10)
		for self.tok != TokEof {
			switch self.tok {
			case TokText:
				r = append(r, printLit(self.lit))
				self.Next()
			case TokTagStart:
				self.Next()
				lit := string(self.lit)
				for _, t := range tags {
					if t == lit {
						self.Next()
						return t, r
					}
				}
				r = append(r, self.parseBlockTag()) // 解析
			case TokVarStart:
				r = append(r, self.parseVarTag())
			default:
				self.Error("unexpected Token %s", self.tok)
			}
		}
		return "", r
	}
*/

// 查找现在silo是否有与父系相同的block.
func (self *Parser) lookup_block(block_name string) *TBlockElement {
	var last_block *TBlockElement // 用于储存最后
	for _, element := range self.silo {
		switch elem := element.(type) {
		case *TBlockElement:
			if elem.name == block_name {
				//log.Print("[lookup_block]:Ture/", block_name, &elem)
				last_block = elem
			}
		}
	}
	return last_block
}

// 主要遍历 block里面的block
func (self *Parser) combine_sublock(sub *TBlockElement) *TBlockElement {
	for i, element := range sub.elements {
		switch elem := element.(type) {
		case *TBlockElement:
			//log.Print("子块查询:", elem.name)
			new_block := self.lookup_block(elem.name) // 查找当前有没有匹配的供使用
			if new_block != nil {
				sub.elements[i] = new_block
			} else {
				sub.elements[i] = self.combine_sublock(elem)
			}
		}
	}
	return sub
}

// 主要遍历模板的block
func (self *Parser) combine_block() {
	for i, element := range self.parent.silo { // 遍历父系所有Block元素
		//log.Print("[combine_block]:element/", element)
		switch elem := element.(type) {
		case *TBlockElement:
			new_block := self.lookup_block(elem.name) // 查找当前有没有匹配的供使用
			if new_block != nil {
				self.parent.silo[i] = new_block
				//log.Print("[combine_block]:TBlockElement/", &element, self.lookup_block(elem.name))
			} else {
				self.parent.silo[i] = self.combine_sublock(elem)
				//log.Print("[空白block]:nil/", element)
			}

		}
	}

}

// 渲染Block,Blcok其实就是TTextElement树
func (self *Parser) render_block(block *TBlockElement, contextChain []interface{}, buf io.Writer) {
	for _, element := range block.elements {
		switch elem := element.(type) {
		case *TBlockElement:
			self.renderElement(elem, contextChain, buf)
		case *TTextElement:
			buf.Write(elem.text) // 无需处理直接返回
		}
	}

}

func (self *Parser) renderTemplate(contextChain []interface{}, buf io.Writer) {
	// 循环对比替换 模板的Tags
	for _, element := range self.silo {
		self.renderElement(element, contextChain, buf) // 渲染各个Element
	}
}

func (self *Parser) renderElement(element interface{}, contextChain []interface{}, buf io.Writer) {
	// 分类处理
	switch elem := element.(type) {
	case *TBlockElement:
		self.render_block(elem, contextChain, buf)
	case *TTextElement:
		buf.Write(elem.text) // 无需处理直接返回
		//log.Println(string(elem.text))
	case *TVarElement:
		break
	case *TSectionElement:
		break
	case *TTemplate:
		break
	default:
		buf.Write([]byte("ddd")) // 无需处理直接返回
	}
}

// Produces a nice error message and returns an error-object.
// The 'token'-argument is optional. If provided, it will take
// the token's position information. If not provided, it will
// automatically use the CURRENT token's position information.
func (p *Parser) Error(msg string, token *TToken) *Error {
	if token == nil {
		// Set current token
		token = p.Current()
		if token == nil {
			// Set to last token
			if len(p.tokens) > 0 {
				token = p.tokens[len(p.tokens)-1]
			}
		}
	}
	var line, col int
	if token != nil {
		line = token.Line
		col = token.Col
	}
	return &Error{
		Template: p.template,
		Filename: p.name,
		Sender:   "parser",
		Line:     line,
		Column:   col,
		Token:    token,
		ErrorMsg: msg,
	}
}

// IDENT if IDENT else IDENT
func (p *Parser) parseExprIfWithElse() (res_eval IEvaluator, res_err *Error) {
	// Parse the variable name
	res_eval, res_err = p.parseVariableOrLiteralWithFilter()
	if res_err != nil {
		return nil, res_err
	}

	// Check for identifier
	if res_eval == nil {
		return nil, p.Error("expression must have an identifier at begin.", nil)
	}

	//	fmt.Println("if TokenKeyword")
	//	# 查找If关键字
	if tokenName := p.Match(TokenKeyword, "if"); tokenName != nil {
		//return nil, p.Error("expression must have an keyword 'if'.", nil)
		// # 创建EvalorEvalor
		node := &nodeIf{}

		// #获取IF表达式
		var argsToken []*TToken
		for p.Peek(TokenKeyword, "else") == nil && p.Remaining() > 1 { //#p.Remaining() > 1 保留1个 }}
			//if p.PeekN(1, TokenSymbol, "}}") != nil {
			//	break
			//}

			// Add token to args
			argsToken = append(argsToken, p.Current())
			p.Consume() // next token

		}

		// # 创建表达式解析器
		argParser := newParser(p.name, argsToken, p.template)
		if len(argsToken) == 0 {
			// This is done to have nice EOF error messages
			argParser.lastToken = tokenName
		}

		if p.template != nil {
			p.template.level++
			defer func() { p.template.level-- }()
		}

		// Parse first and main IF condition
		condition, err := argParser.ParseExpression()
		if err != nil {
			return nil, err
		}
		//		fmt.Println("if condition", condition)

		node.conditions = append(node.conditions, condition) // #条件
		node.wrappers = append(node.wrappers, res_eval)      // #add if value

		// EOF? 表明无else
		if p.Peek(TokenSymbol, "}}") != nil { //if p.Remaining() == 0 {
			//fmt.Println("if Remaining() == 0", p.Current())
			//return nil, p.Error("Unexpectedly reached EOF, no tag end found.", p.lastToken)
			return node, nil
		}

		p.Consume()

		//		fmt.Println("if TokenIdentifier", p.Current())
		res_eval2, err := p.parseVariableOrLiteralWithFilter()
		if err != nil {
			return nil, err
		}

		// add else value
		node.wrappers = append(node.wrappers, res_eval2)

		//	fmt.Println("if else end", p.Current())
		return node, nil

	}

	return res_eval, nil
}

// IDENT | IDENT.(IDENT|NUMBER)...
func (p *Parser) parseVariableOrLiteral() (IEvaluator, *Error) {
	t := p.Current()

	if t == nil {
		return nil, p.Error("Unexpected EOF, expected a number, string, keyword or identifier.", p.lastToken)
	}
	//fmt.Println("parseVariableOrLiteral", t.Typ, TokenNumber, t.Val)
	// Is first part a number or a string, there's nothing to resolve (because there's only to return the value then)
	switch t.Typ {
	case TokenNumber:
		p.Consume()
		//fmt.Println("parseVariableOrLiteral", t.Typ, t.Val)
		// One exception to the rule that we don't have float64 literals is at the beginning
		// of an expression (or a variable name). Since we know we started with an integer
		// which can't obviously be a variable name, we can check whether the first number
		// is followed by dot (and then a number again). If so we're converting it to a float64.

		if p.Match(TokenSymbol, ".") != nil {
			// float64
			t2 := p.MatchType(TokenNumber)
			if t2 == nil {
				return nil, p.Error("Expected a number after the '.'.", nil)
			}
			f, err := strconv.ParseFloat(fmt.Sprintf("%s.%s", t.Val, t2.Val), 64)
			if err != nil {
				return nil, p.Error(err.Error(), t)
			}
			fr := &floatResolver{
				locationToken: t,
				val:           f,
			}
			return fr, nil
		}
		i, err := strconv.Atoi(t.Val)
		if err != nil {
			return nil, p.Error(err.Error(), t)
		}
		nr := &intResolver{
			locationToken: t,
			val:           i,
		}
		return nr, nil

	case TokenString:
		p.Consume()
		sr := &stringResolver{
			locationToken: t,
			val:           t.Val,
		}
		return sr, nil
	case TokenKeyword:
		p.Consume()
		switch t.Val {
		case "true":
			br := &boolResolver{
				locationToken: t,
				val:           true,
			}
			return br, nil
		case "false":
			br := &boolResolver{
				locationToken: t,
				val:           false,
			}
			return br, nil
		default:
			return nil, p.Error("This keyword is not allowed here.", nil)
		}
	}

	resolver := &variableResolver{
		locationToken: t,
	}

	// First part of a variable MUST be an identifier
	if t.Typ != TokenIdentifier {
		return nil, p.Error("Expected either a number, string, keyword or identifier.", t)
	}

	resolver.parts = append(resolver.parts, &variablePart{
		typ: varTypeIdent,
		s:   t.Val,
	})

	p.Consume() // we consumed the first identifier of the variable name

variableLoop:
	for p.Remaining() > 0 {
		t = p.Current()

		if p.Match(TokenSymbol, ".") != nil {
			// Next variable part (can be either NUMBER or IDENT)
			t2 := p.Current()
			if t2 != nil {
				switch t2.Typ {
				case TokenIdentifier:
					resolver.parts = append(resolver.parts, &variablePart{
						typ: varTypeIdent,
						s:   t2.Val,
					})
					p.Consume() // consume: IDENT
					continue variableLoop
				case TokenNumber:
					i, err := strconv.Atoi(t2.Val)
					if err != nil {
						return nil, p.Error(err.Error(), t2)
					}
					resolver.parts = append(resolver.parts, &variablePart{
						typ: varTypeInt,
						i:   i,
					})
					p.Consume() // consume: NUMBER
					continue variableLoop
				default:
					return nil, p.Error("This token is not allowed within a variable name.", t2)
				}
			} else {
				// EOF
				return nil, p.Error("Unexpected EOF, expected either IDENTIFIER or NUMBER after DOT.",
					p.lastToken)
			}
		} else if p.Match(TokenSymbol, "(") != nil {
			// Function call
			// FunctionName '(' Comma-separated list of expressions ')'
			part := resolver.parts[len(resolver.parts)-1]
			part.isFunctionCall = true
		argumentLoop:
			for {
				if p.Remaining() == 0 {
					return nil, p.Error("Unexpected EOF, expected function call argument list.", p.lastToken)
				}

				if p.Peek(TokenSymbol, ")") == nil {
					// No closing bracket, so we're parsing an expression
					exprArg, err := p.ParseExpression()
					if err != nil {
						return nil, err
					}
					part.callingArgs = append(part.callingArgs, exprArg)

					if p.Match(TokenSymbol, ")") != nil {
						// If there's a closing bracket after an expression, we will stop parsing the arguments
						break argumentLoop
					} else {
						// If there's NO closing bracket, there MUST be an comma
						if p.Match(TokenSymbol, ",") == nil {
							return nil, p.Error("Missing comma or closing bracket after argument.", nil)
						}
					}
				} else {
					// We got a closing bracket, so stop parsing arguments
					p.Consume()
					break argumentLoop
				}

			}
			// We're done parsing the function call, next variable part
			continue variableLoop
		}

		// No dot or function call? Then we're done with the variable parsing
		break
	}

	return resolver, nil
}

func (p *Parser) parseVariableOrLiteralWithFilter() (*nodeFilteredVariable, *Error) {
	v := &nodeFilteredVariable{
		locationToken: p.Current(),
	}

	// Parse the variable name
	resolver, err := p.parseVariableOrLiteral()
	if err != nil {
		return nil, err
	}
	v.resolver = resolver

	// Parse all the filters
filterLoop:
	for p.Match(TokenSymbol, "|") != nil {
		// Parse one single filter
		filter, err := p.parseFilter()
		if err != nil {
			return nil, err
		}

		// Check sandbox filter restriction
		if _, isBanned := p.template.set.bannedFilters[filter.name]; isBanned {
			return nil, p.Error(fmt.Sprintf("Usage of filter '%s' is not allowed (sandbox restriction active).", filter.name), nil)
		}

		v.filterChain = append(v.filterChain, filter)

		continue filterLoop

	}

	return v, nil
}

func (p *Parser) parseVariableElement() (INode, *Error) {
	node := &nodeVariable{
		locationToken: p.Current(),
	}

	p.Consume() // consume '{{'

	expr, err := p.ParseExpression() //ParseExpression("{{")
	if err != nil {
		return nil, err
	}

	//fmt.Println()
	node.expr = expr

	if p.Match(TokenSymbol, "}}") == nil {
		return nil, p.Error("'}}' expected", nil)
	}

	return node, nil
}

func (p *Parser) ParseDocument(content string) (res *nodeDocument, res_err *Error) {
	p.tokens, res_err = NewLexer(p.name, html.UnescapeString(content))
	if res_err != nil {
		return nil, res_err
	}

	return p.parseDocument()
}

func (p *Parser) parseDocument() (*nodeDocument, *Error) {
	doc := &nodeDocument{}

	for p.Remaining() > 0 {
		node, err := p.parseDocElement()
		if err != nil {
			return nil, err
		}

		//fmt.Println("node ", reflect.TypeOf(node))
		doc.Nodes = append(doc.Nodes, node)
	}

	return doc, nil
}

// Wraps all nodes between starting tag and "{% endtag %}" and provides
// one simple interface to execute the wrapped nodes.
// It returns a parser to process provided arguments to the tag.
func (p *Parser) WrapUntilTag(names ...string) (*NodeWrapper, *Parser, *Error) {
	wrapper := &NodeWrapper{}

	var tagArgs []*TToken

	for p.Remaining() > 0 {
		// New tag, check whether we have to stop wrapping here
		if p.Peek(TokenSymbol, "{%") != nil {
			tagIdent := p.PeekTypeN(1, TokenIdentifier)

			if tagIdent != nil {
				// We've found a (!) end-tag

				found := false
				for _, n := range names {
					if tagIdent.Val == n {
						found = true
						break
					}
				}

				// We only process the tag if we've found an end tag
				if found {
					// Okay, endtag found.
					p.ConsumeN(2) // '{%' tagname

					for {
						if p.Match(TokenSymbol, "%}") != nil {
							// Okay, end the wrapping here
							wrapper.Endtag = tagIdent.Val
							name := "x"
							if p.template != nil {
								name = p.template.Name
							}
							return wrapper, newParser(name, tagArgs, p.template), nil
						}
						t := p.Current()
						p.Consume()
						if t == nil {
							return nil, nil, p.Error("Unexpected EOF.", p.lastToken)
						}
						tagArgs = append(tagArgs, t)
					}
				}
			}

		}

		// Otherwise process next element to be wrapped
		node, err := p.parseDocElement()
		if err != nil {
			return nil, nil, err
		}
		wrapper.nodes = append(wrapper.nodes, node)
	}

	return nil, nil, p.Error(fmt.Sprintf("Unexpected EOF, expected tag %s.", strings.Join(names, " or ")),
		p.lastToken)
}

// Returns the total token count.
func (p *Parser) Count() int {
	return len(p.tokens)
}

// Tag = "{%" IDENT ARGS "%}"
func (p *Parser) parseTagElement() (INodeTag, *Error) {
	p.Consume() // consume "{%"
	tokenName := p.MatchType(TokenIdentifier)

	// Check for identifier
	if tokenName == nil {
		return nil, p.Error("Tag name must be an identifier.", nil)
	}

	// Check for the existing tag
	tag, exists := tags[tokenName.Val]
	if !exists {
		// Does not exists
		return nil, p.Error(fmt.Sprintf("Tag '%s' not found (or beginning tag not provided)", tokenName.Val), tokenName)
	}

	// Check sandbox tag restriction
	if p.template != nil {
		if _, isBanned := p.template.set.bannedTags[tokenName.Val]; isBanned {
			return nil, p.Error(fmt.Sprintf("Usage of tag '%s' is not allowed (sandbox restriction active).", tokenName.Val), tokenName)
		}
	}
	var argsToken []*TToken
	for p.Peek(TokenSymbol, "%}") == nil && p.Remaining() > 0 {
		// Add token to args
		argsToken = append(argsToken, p.Current())
		p.Consume() // next token
	}

	// EOF?
	if p.Remaining() == 0 {
		return nil, p.Error("Unexpectedly reached EOF, no tag end found.", p.lastToken)
	}

	p.Match(TokenSymbol, "%}")

	argParser := newParser(p.name, argsToken, p.template)
	if len(argsToken) == 0 {
		// This is done to have nice EOF error messages
		argParser.lastToken = tokenName
	}

	if p.template != nil {
		p.template.level++
		defer func() { p.template.level-- }()
	}

	return tag.parser(p, tokenName, argParser)
}

// Doc = { ( Filter | Tag | HTML ) }
func (p *Parser) parseDocElement() (INode, *Error) {
	t := p.Current()

	switch t.Typ {
	case TokenHTML: // #文本内容
		p.Consume() // consume HTML element
		return &nodeHTML{token: t}, nil
	case TokenSymbol: // #目标内容
		switch t.Val {
		case "{{":
			// parse variable
			variable, err := p.parseVariableElement()
			if err != nil {
				return nil, err
			}
			return variable, nil
		case "{%":
			// parse tag
			tag, err := p.parseTagElement()
			if err != nil {
				return nil, err
			}
			return tag, nil
		}
	}
	return nil, p.Error("Unexpected token (only HTML/tags/filters in templates allowed)", t)
}

func (p *Parser) parseFactor() (res_eval IEvaluator, res_err *Error) {
	if p.Match(TokenSymbol, "(") != nil {
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		if p.Match(TokenSymbol, ")") == nil {
			return nil, p.Error("Closing bracket expected after expression", nil)
		}
		return expr, nil
	}

	/*
		res_eval, res_err = p.parseVariableOrLiteralWithFilter()

		if p.Match(TokenKeyword, "if") != nil {
			node := &nodeIf{}node := &nodeIf{}
			node.wrappers=append(node.wrappers,res_eval)
			fmt.Println("if start")
			res_eval, res_err = p.parseExprIfWithElse(node)
		}
	*/

	return p.parseExprIfWithElse()
}

func (p *Parser) parsePower() (IEvaluator, *Error) {
	pw := new(power)

	power1, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	pw.power1 = power1

	if p.Match(TokenSymbol, "^") != nil {
		power2, err := p.parsePower()
		if err != nil {
			return nil, err
		}
		pw.power2 = power2
	}

	if pw.power2 == nil {
		// Shortcut for faster evaluation
		return pw.power1, nil
	}

	return pw, nil
}

func (p *Parser) parseTerm() (IEvaluator, *Error) {
	returnTerm := new(term)

	factor1, err := p.parsePower()
	if err != nil {
		return nil, err
	}
	returnTerm.factor1 = factor1

	for p.PeekOne(TokenSymbol, "*", "/", "%") != nil {
		if returnTerm.operator != nil {
			// Create new sub-term
			returnTerm = &term{
				factor1: returnTerm,
			}
		}

		op := p.Current()
		p.Consume()

		factor2, err := p.parsePower()
		if err != nil {
			return nil, err
		}

		returnTerm.operator = op
		returnTerm.factor2 = factor2
	}

	if returnTerm.operator == nil {
		// Shortcut for faster evaluation
		return returnTerm.factor1, nil
	}

	return returnTerm, nil
}

func (p *Parser) parseSimpleExpression() (IEvaluator, *Error) {
	expr := new(simpleExpression)

	if sign := p.MatchOne(TokenSymbol, "+", "-"); sign != nil {
		if sign.Val == "-" {
			expr.negativeSign = true
		}
	}

	if p.Match(TokenSymbol, "!") != nil || p.Match(TokenKeyword, "not") != nil {
		expr.negate = true
	}

	term1, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	expr.term1 = term1

	for p.PeekOne(TokenSymbol, "+", "-") != nil {
		if expr.operator != nil {
			// New sub expr
			expr = &simpleExpression{
				term1: expr,
			}
		}

		op := p.Current()
		p.Consume()

		term2, err := p.parseTerm()
		if err != nil {
			return nil, err
		}

		expr.term2 = term2
		expr.operator = op
	}

	if expr.negate == false && expr.negativeSign == false && expr.term2 == nil {
		// Shortcut for faster evaluation
		return expr.term1, nil
	}

	return expr, nil
}

func (p *Parser) parseRelationalExpression() (IEvaluator, *Error) {
	expr1, err := p.parseSimpleExpression()
	if err != nil {
		return nil, err
	}

	expr := &relationalExpression{
		expr1: expr1,
	}

	if t := p.MatchOne(TokenSymbol, "==", "<=", ">=", "!=", "<>", ">", "<"); t != nil {
		expr2, err := p.parseRelationalExpression()
		if err != nil {
			return nil, err
		}
		expr.operator = t
		expr.expr2 = expr2
	} else if t := p.MatchOne(TokenKeyword, "in"); t != nil {
		expr2, err := p.parseSimpleExpression()
		if err != nil {
			return nil, err
		}
		expr.operator = t
		expr.expr2 = expr2
	}

	if expr.expr2 == nil {
		// Shortcut for faster evaluation
		return expr.expr1, nil
	}

	return expr, nil
}

// #解析表达式 TODO 非Export
func (p *Parser) ParseExpression(eva ...IEvaluator) (IEvaluator, *Error) {
	var rexpr1 IEvaluator
	var err *Error
	if len(eva) == 0 {
		rexpr1, err = p.parseRelationalExpression()
		if err != nil {
			return nil, err
		}
	} else {
		rexpr1 = eva[0]
	}

	exp := &TExpression{
		expr1: rexpr1,
		//		tag:   ltag,
	}
	//fmt.Println("ParseExpression 1", reflect.TypeOf(rexpr1))

	// # 以表达式关键字为界拆分表达式
	if p.PeekOne(TokenSymbol, "&&", "||") != nil || p.PeekOne(TokenKeyword, "and", "or", "if", "else") != nil {
		op := p.Current()
		p.Consume()
		expr2, err := p.parseRelationalExpression()
		if err != nil {
			return nil, err
		}
		exp.expr2 = expr2
		exp.operator = op
		//fmt.Println("ParseExpression 2", op.String(), reflect.TypeOf(expr2))
	}

	if exp.expr2 == nil {
		// Shortcut for faster evaluation
		return exp.expr1, nil
	}

	return p.ParseExpression(exp)
}

// Returns tokens[i] or NIL (if i >= len(tokens))
func (p *Parser) Get(i int) *TToken {
	if i < len(p.tokens) {
		//fmt.Println("Get", i, p.tokens[i])
		return p.tokens[i]
	}
	return nil
}

// Consume one token. It will be gone forever.
func (p *Parser) Consume() {
	p.ConsumeN(1)
}

// Consume N tokens. They will be gone forever.
func (p *Parser) ConsumeN(count int) {
	p.idx += count
}

// Returns the UNCONSUMED token count.
func (p *Parser) Remaining() int {
	return len(p.tokens) - p.idx
}

// Returns the current token.
func (p *Parser) Current() *TToken {
	return p.Get(p.idx)
}

// Returns the CURRENT token if the given type AND value matches.
// Consumes this token on success.
func (p *Parser) Match(typ TokenType, val string) *TToken {
	if t := p.Peek(typ, val); t != nil {
		p.Consume()
		return t
	}
	return nil
}

// Returns the CURRENT token if the given type AND value matches.
// It DOES NOT consume the token.
func (p *Parser) Peek(typ TokenType, val string) *TToken {
	return p.PeekN(0, typ, val)
}

// Returns the CURRENT token if the given type AND *one* of
// the given values matches.
// It DOES NOT consume the token.
func (p *Parser) PeekOne(typ TokenType, vals ...string) *TToken {
	for _, v := range vals {
		t := p.PeekN(0, typ, v)
		if t != nil {
			return t
		}
	}
	return nil
}

// Returns the CURRENT token if the given type matches.
// It DOES NOT consume the token.
func (p *Parser) PeekType(typ TokenType) *TToken {
	return p.PeekTypeN(0, typ)
}

// Returns the tokens[current position + shift] token if the given type matches.
// DOES NOT consume the token for that token.
func (p *Parser) PeekTypeN(shift int, typ TokenType) *TToken {
	t := p.Get(p.idx + shift)
	if t != nil {
		//fmt.Println("PeekTypeN", t)
		if t.Typ == typ {
			return t
		}
	}
	return nil
}

// Returns the CURRENT token if the given type matches.
// Consumes this token on success.
func (p *Parser) MatchType(typ TokenType) *TToken {
	if t := p.PeekType(typ); t != nil {
		p.Consume()
		return t
	}
	return nil
}

// Returns the tokens[current position + shift] token if the
// given type AND value matches for that token.
// DOES NOT consume the token.
func (p *Parser) PeekN(shift int, typ TokenType, val string) *TToken {
	t := p.Get(p.idx + shift)
	if t != nil {
		//fmt.Println("PeekN", t)
		if t.Typ == typ && t.Val == val {
			return t
		}
	}
	return nil
}

// Returns the CURRENT token if the given type AND *one* of
// the given values matches.
// Consumes this token on success.
func (p *Parser) MatchOne(typ TokenType, vals ...string) *TToken {
	for _, val := range vals {
		if t := p.Peek(typ, val); t != nil {
			p.Consume()
			return t
		}
	}
	return nil
}
