package template

import (
	"fmt"
	"strings"
	"unicode/utf8"
	//	"bytes"
	//	"unicode"
	//	"unicode/utf8"
	"io"
)

const (
	TokenError = iota
	EOF

	TokenHTML

	TokenKeyword
	TokenIdentifier
	TokenString
	TokenNumber
	TokenSymbol
)

var (
	VarLeftDelim  = "{{" //开始的标志
	VarRightDelim = "}}" //结束的标志
	TagLeftDelim  = "{%" //开始的标志
	TagRightDelim = "%}" //结束的标志

	tokenSpaceChars                = " \n\r\t"
	tokenIdentifierChars           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
	tokenIdentifierCharsWithDigits = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789"
	tokenDigits                    = "0123456789"

	// Available symbols in pongo2 (within filters/tag)
	TokenSymbols = []string{
		// 3-Char symbols

		// 2-Char symbols
		"==", ">=", "<=", "&&", "||", "{{", "}}", "{%", "%}", "!=", "<>",

		// 1-Char symbol
		"(", ")", "+", "-", "*", "<", ">", "/", "^", ",", ".", "!", "|", ":", "=", "%",
	}

	// Available keywords in pongo2
	TokenKeywords = []string{"in", "and", "or", "not", "true", "false", "as", "export", "if", "else"}
)

/*
type Token int

const (
	TokIllegal Token = iota
	TokEof

	TokIdent  // for
	TokInt    // 12345
	TokFloat  // 123.45e2
	TokString // 'abc' or "abc"

	TokText // text not inside a tag

	TokTagStart // {%
	TokTagEnd   // %}
	TokVarStart // {{
	TokVarEnd   // }}

	TokAdd // +
	TokSub // -
	TokMul // *
	TokDiv // /
	TokRem // %

	TokAnd // and
	TokOr  // or
	TokNot // not

	TokEqual     // ==
	TokLess      // <
	TokLessEq    // <=
	TokGreater   // >
	TokGreaterEq // >=
	TokNotEq     // !=

	TokDot   // .
	TokBar   // |
	TokColon // :
)

// 只提供名称给Token.String 调用
var tokStrings = map[Token]string{
	TokIllegal:   "illegal",
	TokEof:       "eof",
	TokIdent:     "ident",
	TokInt:       "int",
	TokFloat:     "float",
	TokString:    "string",
	TokText:      "text",
	TokTagStart:  "{%",
	TokTagEnd:    "%}",
	TokVarStart:  "{{",
	TokVarEnd:    "}}",
	TokAdd:       "+",
	TokSub:       "-",
	TokMul:       "*",
	TokDiv:       "/",
	TokRem:       "%",
	TokAnd:       "and",
	TokOr:        "or",
	TokNot:       "not",
	TokEqual:     "==",
	TokLess:      "<",
	TokLessEq:    "<=",
	TokGreater:   ">",
	TokGreaterEq: ">=",
	TokNotEq:     "!=",
	TokDot:       ".",
	TokBar:       "|",
	TokColon:     ":",
}

func (t Token) String() string {
	return tokStrings[t]
}

func (t Token) Precedence() int {
	switch t {
	case TokOr:
		return 1
	case TokAnd:
		return 2
	case TokEqual, TokNotEq, TokLess, TokLessEq, TokGreater, TokGreaterEq:
		return 3
	case TokAdd, TokSub:
		return 4
	case TokMul, TokDiv, TokRem:
		return 5
	}
	return 0
}
*/

// lexer提供数据的扫描工作
type (
	TokenType int

	TToken struct {
		Filename string
		Typ      TokenType
		Val      string
		Line     int
		Col      int
	}

	lexerStateFn func() lexerStateFn
	lexer        struct {
		name       string
		input      string //即将被扫描的文件流 一般放在执行扫描的类里
		leftdelim  string //开始的标志
		rightdelim string //结束的标志
		width      int    // width of last rune
		start      int
		pos        int // 当前游标

		col          int // 某行第几个
		line         int // 行
		startline    int
		startcol     int
		tokens       []*TToken // 存储节点
		curline      int
		errored      bool
		inVerbatim   bool
		verbatimName string
		/*
			begin     int //从这里开始
			end       int //结束的地方
			offset    int // current position in the input.
			ch        rune
			width     int // width of last rune read from input.
			insideTag bool
		*/
	}
)

func NewLexer(name string, input string) ([]*TToken, *Error) {
	l := &lexer{
		name:      name,
		input:     input,
		tokens:    make([]*TToken, 0, 100),
		line:      1,
		col:       1,
		startline: 1,
		startcol:  1,
	}

	l.run()
	if l.errored {
		errtoken := l.tokens[len(l.tokens)-1]
		return nil, &Error{
			Filename: name,
			Line:     errtoken.Line,
			Column:   errtoken.Col,
			Sender:   "lexer",
			ErrorMsg: errtoken.Val,
		}
	}
	return l.tokens, nil
}

func (t *TToken) String() string {
	val := t.Val
	if len(val) > 1000 {
		val = fmt.Sprintf("%s...%s", val[:10], val[len(val)-5:len(val)])
	}

	typ := ""
	switch t.Typ {
	case TokenHTML:
		typ = "HTML"
	case TokenError:
		typ = "Error"
	case TokenIdentifier:
		typ = "Identifier"
	case TokenKeyword:
		typ = "Keyword"
	case TokenNumber:
		typ = "Number"
	case TokenString:
		typ = "String"
	case TokenSymbol:
		typ = "Symbol"
	default:
		typ = "Unknown"
	}

	return fmt.Sprintf("<Token Typ=%s (%d) Val='%s' Line=%d Col=%d>",
		typ, t.Typ, val, t.Line, t.Col)
}

/*
func (self *lexer) init() {
	self.begin=0
	self.end=0
	self.pos=0
	self.ch, self.width = utf8.DecodeRune(self.buf) //返回文件第一个字符
}

func (self *lexer)harvest(){}

// 扫描数据
func (self *lexer) scan() (string, Block) {
loop:
	// 如果现有条件不足直接调用 scanText 查找满足的条件
	if !self.insideTag && self.ch != '{' {

		return self.scanText()
	}
	self.insideTag = true
	self.consumeWhitespace() //删除空白

	pos := self.offset
	tok := TokIllegal

	switch ch := self.ch; {
	case unicode.IsLetter(ch), self.ch == '_':
		return self.scanIdent()
	case unicode.IsDigit(ch):
		tok = self.scanNumber()
	case ch == '|':
		tok = TokBar
		self.next()
	case ch == '.':
		tok = TokDot
		self.next()
	case ch == ':':
		tok = TokColon
		self.next()
	case ch == '+':
		tok = TokAdd
		self.next()
	case ch == '-':
		tok = TokSub
		self.next()
	case ch == '*':
		tok = TokMul
		self.next()
	case ch == '/':
		tok = TokDiv
		self.next()
	default:
		self.next()
		switch ch {
		case -1:
			tok = TokEof
		case '{':
			switch self.ch {
			case '%':
				tok = TokTagStart
				self.next()
			case '{':
				tok = TokVarStart
				self.next()
			case '#':
				// start of a comment; scan until the end
				self.next()
				for {
					for self.ch != '#' {
						self.next()
					}
					self.next()
					if self.ch == '}' {
						self.next()
						break
					}
				}
				goto scanAgain
			}
		case '%':
			tok = TokRem
			if self.ch == '}' {
				self.insideTag = false
				tok = TokTagEnd
				self.next()
			}
		case '}':
			if self.ch == '}' {
				self.insideTag = false
				tok = TokVarEnd
				self.next()
			}
		case '\'', '"':
			tok = self.scanString(byte(ch))
			return tok, self.buf[pos+1 : self.offset-1]
		case '<':
			tok = TokLess
			if self.ch == '=' {
				tok = TokLessEq
				self.next()
			}
		case '>':
			tok = TokGreater
			if self.ch == '=' {
				tok = TokGreaterEq
				self.next()
			}
		case '=':
			if self.ch == '=' {
				tok = TokEqual
				self.next()
			}
		case '!':
			if self.ch == '=' {
				tok = TokNotEq
				self.next()
			}
		}
	}
	return tok, self.buf[pos:self.offset]
}
*/
/*
//主要-略过特殊字符移动
func (self *lexer) consume_whitespace() int {
	var count=0
	for self.ch == ' ' || self.ch == '\t' || self.ch == '\n' || self.ch == '\r' {
		self.next()
		count++
	}

	return count
}
*/

// #重置
func (l *lexer) rest() {
	l.name = ""
	l.input = ""
	l.tokens = make([]*TToken, 0, 100)
	l.line = 1
	l.col = 1
	l.startline = 1
	l.startcol = 1
	l.errored = false
}

func (l *lexer) run() {
	for {
		// TODO: Support verbatim tag names
		// https://docs.djangoproject.com/en/dev/ref/templates/builtins/#verbatim
		if l.inVerbatim {
			name := l.verbatimName
			if name != "" {
				name += " "
			}
			if strings.HasPrefix(l.input[l.pos:], fmt.Sprintf("{%% endverbatim %s%%}", name)) { // end verbatim
				if l.pos > l.start {
					l.emit(TokenHTML)
				}
				w := len("{% endverbatim %}")
				l.pos += w
				l.col += w
				l.ignore()
				l.inVerbatim = false
			}
		} else if strings.HasPrefix(l.input[l.pos:], "{% verbatim %}") { // tag
			if l.pos > l.start {
				l.emit(TokenHTML)
			}
			l.inVerbatim = true
			w := len("{% verbatim %}")
			l.pos += w
			l.col += w
			l.ignore()
		}

		if !l.inVerbatim {
			// Ignore single-line comments {# ... #}
			if strings.HasPrefix(l.input[l.pos:], "{#") {
				if l.pos > l.start {
					l.emit(TokenHTML)
				}

				l.pos += 2 // pass '{#'
				l.col += 2

				for {
					switch l.peek() {
					case EOF:
						l.errorf("Single-line comment not closed.")
						return
					case '\n':
						l.errorf("Newline not permitted in a single-line comment.")
						return
					}

					if strings.HasPrefix(l.input[l.pos:], "#}") {
						l.pos += 2 // pass '#}'
						l.col += 2
						break
					}

					l.next()
				}
				l.ignore() // ignore whole comment

				// Comment skipped
				continue // next token
			}

			if strings.HasPrefix(l.input[l.pos:], "{{") || // variable
				strings.HasPrefix(l.input[l.pos:], "{%") { // tag
				if l.pos > l.start {
					l.emit(TokenHTML)
				}
				l.tokenize()
				if l.errored {
					return
				}
				continue
			}
		}

		switch l.peek() {
		case '\n':
			l.line++
			l.col = 0
		}
		if l.next() == EOF {
			break
		}
	}

	if l.pos > l.start {
		l.emit(TokenHTML)
	}

	if l.inVerbatim {
		l.errorf("verbatim-tag not closed, got EOF.")
	}
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}
func (l *lexer) acceptRun(what string) {
	for strings.IndexRune(what, l.next()) >= 0 {
	}
	l.backup()
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}
func (l *lexer) length() int {
	return l.pos - l.start
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

//主要-移动偏移
// next returns the next rune in the input.
func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	l.col += l.width
	return r
}

func (self *lexer) eof() bool {
	return self.pos >= len(self.input)-1
}

func (l *lexer) errorf(format string, args ...interface{}) lexerStateFn {
	t := &TToken{
		Filename: l.name,
		Typ:      TokenError,
		Val:      fmt.Sprintf(format, args...),
		Line:     l.startline,
		Col:      l.startcol,
	}
	l.tokens = append(l.tokens, t)
	l.errored = true
	l.startline = l.line
	l.startcol = l.col
	return nil
}

func (l *lexer) value() string {
	return l.input[l.start:l.pos]
}

func (l *lexer) tokenize() {
	for state := l.stateCode; state != nil; {
		state = state()
	}
}

func (l *lexer) emit(t TokenType) {
	tok := &TToken{
		Filename: l.name,
		Typ:      t,
		Val:      l.value(),
		Line:     l.startline,
		Col:      l.startcol,
	}

	if t == TokenString {
		// Escape sequence \" in strings
		tok.Val = strings.Replace(tok.Val, `\"`, `"`, -1)
		tok.Val = strings.Replace(tok.Val, `\\`, `\`, -1)
	}

	l.tokens = append(l.tokens, tok)
	l.start = l.pos
	l.startline = l.line
	l.startcol = l.col
}

// 判断扫描 {{
/*
	循环匹配 {{ 直到找到完全匹配的
*/
func (self *lexer) scan_string(s string) (string, error) { // s={{
	cur_pos := self.pos //保存当前 游标
	newlines := 0
	for true { //循环直到匹配到"{{"返回

		if cur_pos+len(s) > len(string(self.input)) { //是否已经到末端了??
			return string(self.input[self.pos:]), io.EOF //如果是-->获取当前游标以后的数据
		}

		if self.input[cur_pos] == '\n' { //判断是否到行末
			newlines++
		}

		// 直到找到 第一个 和S 要查询的字符一样
		if self.input[cur_pos] != s[0] { //当前是不是 目标字符
			cur_pos++ /* 1 offset 1 */
			continue  //重新循环
		}

		//一切判断结束,这里表示匹配到类似字符
		// 表为找到 头个{ 下面匹配另一个{ 如果另一个不是想要的重新循环
		match := true
		for j := 1; j < len(s); j++ { // 从1开始是因为0代表{ 1代表另一个{
			if s[j] != self.input[cur_pos+j] { // 是否匹配成功
				match = false
				break
			}
		}

		// 如果已经匹配到{{
		if match {
			e := cur_pos + len(s)          //偏移到包含 S之内的 游标地址
			text := self.input[self.pos:e] //读取{{这段数据
			self.pos = e                   //设置偏移为当前偏移					/* 2 offset n */

			self.curline += newlines // 记录本次循环到几行
			//log.Print("[LEXER]:",string(text))
			return string(text), nil
		} else {
			// 没找到的话继续循环,这是转向下一个字符的唯一代码
			cur_pos++
		}
	} //for

	//should never be here
	return "", nil
}

/*
//查找满足的条件
func (self *lexer) scanText()(string,Block) {
	// 保存状态
	n := len(self.buf) - 1
	off := self.offset
	beginpos,endpos := 0,0
loop:
	for {
		if self.leftDelim[0] != '{' {
			//________________________________________
								// 当前偏移起
			beginpos = bytes.IndexByte(self.buf[off:], '{') // 定位 { 位置
			if pos < 0 || pos >= n {  // 如果-1 没找到 或 偏移大于整个内存
				break
			}
			off = beginpos // 重置偏移 偏移从0算起
			//________________________________________
			switch self.buf[off+1] {
			case '%':

				self.leftDelim='{%' // 保存
				if self.buf[off+1] !=' '{ // {% extends "xx.html" %} 必须有空白间隔
					self.next
					consume_whitespace()
				}

				if HasPrefix(self.buf[self.offset:],[]byte("extends")){

					endpos=bytes.IndexByte(self.buf[self.offset:], '*}') //
					self.block.leftDelim=`{%`
					self.block.rightDelim=`%}`
					self.block.action=self.buf[self.offset:len("extends")]

					self.block.actionName=Trim(Trim([]byte(self.buf[offset:endpos]),` `),`"`)
					self.block.offset=beginpos
					self.block.length=self.offset-beginpos+endpos+2 //当前-开始编译+相对偏移=长度 2是*}

					return self.block.action, self.block
					}
				if HasPrefix(self.buf[offset:],[]byte(`block`)){
				}

				// 以上都不是那么不是想要的重置所以

				loop

				}
				if HasPrefix(self.buf[offset:],[]byte("block")){}
			case '{':
			case '#':// 有标志符迹象
			default:// 如果恕不标志符则重新扫描定位
				break loop
		}
	self.insideTag = true
	return "",nil
}
/*
func (self *lexer) scanIdent() (Token, []byte) {
	pos := self.offset
	for unicode.IsLetter(self.ch) || unicode.IsDigit(self.ch) || self.ch == '_' {
		self.next()
	}
	lit := self.buf[pos:self.offset]
	tok := TokIdent
	switch string(lit) {
	case "and":
		tok = TokAnd
	case "or":
		tok = TokOr
	case "not":
		tok = TokNot
	}
	return tok, lit
}

func (self *lexer) scanNumber() Token {
	tok := TokInt
	seenDecimal := false
	seenExponent := false

	if self.ch == '-' || self.ch == '+' {
		self.next()
	}
	for self.ch >= '0' && self.ch <= '9' ||
		!seenDecimal && !seenExponent && self.ch == '.' ||
		!seenExponent && (self.ch == 'e' || self.ch == 'E') {
		if self.ch == '.' {
			seenDecimal = true
			tok = TokFloat
		}
		if self.ch == 'e' || self.ch == 'E' {
			seenExponent = true
			tok = TokFloat
		}
		self.next()
	}
	return tok
}

func (self *lexer) scanString(start byte) Token {
	// ' or " already consumed
	pos := bytes.IndexByte(self.buf[self.offset:], start)
	if pos < 0 {
		panic("String not terminated")
	}
	self.offset += pos
	self.width = 1
	self.next()
	return TokString
}
*/

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) stateCode() lexerStateFn {
outer_loop:
	for {
		switch {
		case l.accept(tokenSpaceChars):
			if l.value() == "\n" {
				return l.errorf("Newline not allowed within tag/variable.")
			}
			l.ignore()
			continue
		case l.accept(tokenIdentifierChars):
			return l.stateIdentifier
		case l.accept(tokenDigits):
			return l.stateNumber
		case l.accept(`'`) || l.accept(`"`):
			return l.stateString
		}

		// Check for symbol
		for _, sym := range TokenSymbols {
			if strings.HasPrefix(l.input[l.start:], sym) {
				l.pos += len(sym)
				l.col += l.length()
				l.emit(TokenSymbol)

				if sym == "%}" || sym == "}}" {
					// Tag/variable end, return after emit
					return nil
				}

				continue outer_loop
			}
		}

		break
	}

	// Normal shut down
	return nil
}

func (l *lexer) stateIdentifier() lexerStateFn {
	l.acceptRun(tokenIdentifierChars)
	l.acceptRun(tokenIdentifierCharsWithDigits)
	for _, kw := range TokenKeywords {
		if kw == l.value() {
			l.emit(TokenKeyword)
			return l.stateCode
		}
	}
	l.emit(TokenIdentifier)
	return l.stateCode
}

func (l *lexer) stateNumber() lexerStateFn {
	l.acceptRun(tokenDigits)
	/*
		Maybe context-sensitive number lexing?
		* comments.0.Text // first comment
		* usercomments.1.0 // second user, first comment
		* if (score >= 8.5) // 8.5 as a number

		if l.peek() == '.' {
			l.accept(".")
			if !l.accept(tokenDigits) {
				return l.errorf("Malformed number.")
			}
			l.acceptRun(tokenDigits)
		}
	*/
	l.emit(TokenNumber)
	return l.stateCode
}

func (l *lexer) stateString() lexerStateFn {
	l.ignore()
	l.startcol-- // we're starting the position at the first "
	//fmt.Println("stateString", l.accept(`'`), l.accept(`"`))
	for !l.accept(`'`) && !l.accept(`"`) {
		switch l.next() {
		case '\\':
			// escape sequence
			switch l.peek() {
			case '\'', '"', '\\':
				l.next()
			default:
				return l.errorf("Unknown escape sequence: \\%c", l.peek())
			}
		case EOF:
			return l.errorf("Unexpected EOF, string not closed.")
		case '\n':
			return l.errorf("Newline in string is not allowed.")
		}
	}
	l.backup()
	l.emit(TokenString)

	l.next()
	l.ignore()

	return l.stateCode
}
