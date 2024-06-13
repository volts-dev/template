package template

import (
	"bytes"
	"html/template"
	"io"
	"reflect"
	"time"
)

/*
var tags = map[string]ParseFunc {
	"block":     parseBlock,
	"extends":   parseExtends,

	"cycle":     parseCycle,
	"firstof":   parseFirstof,
	"for":       parseFor,
	"if":        parseIf,
	"ifchanged": parseIfChanged,
	"include":   parseInclude,
	"override":  parseOverride,
	"set":       parseSet,
	"with":      parseWith,

}
*/
// Template is the representation of a parsed template. The *parse.Tree
// field is exported only for use by html/template and should be treated
// as unexported by all other clients.
// Template 用来封装Parser 提供的行数 来提供对外功能
type (
	// 模板资源文件
	TAssetFile struct {
		Name     string // 文件名称
		Path     string // 路径
		Type     string // js,css,html
		Content  string
		Media    string
		Modified time.Time // 修改时间
	}

	TTemplate struct {
		Engine   IEngine
		Name     string
		DirPath  string // 模版文件夹路径
		parser   *Parser
		template *template.Template
		langMap  map[string]string

		level int
	}
)

// Version string
const Version = "beta"

var (
	//tplcache cacher.ICache
	TemplateFuncs = map[string]interface{}{
		"trans": func(aText string) template.HTML {
			return template.HTML(aText)
			// 在Router里实现了！
		},
		"Str2html": Str2html,
	}
	TemplateVars = map[string]interface{}{} //储存[模板变量]

	defaultTemplateSet = New()
)

// 创建一个Template
func NewTemplate(aName string) *TTemplate {
	/*
		//    cwd := os.Getenv("CWD")
		// 创建词法分析器
		xlexer := &lexer{
			stream:     htmlstream,
			leftdelim:  LEFT_DELIM,
			rightdelim: RIGHT_DELIM,
			pos:        0,
			curline:    1}

		// 创建分析器
		xparser := &Parser{
			templete: self,
			parent:   nil,
			lex:      xlexer,
			silo:     []interface{}{}}
	*/
	lTmpl := &TTemplate{
		Name: aName,

		//langMap: langmap,
	} //创建一个新的Template
	lTmpl.parser = NewTemplateParser(lTmpl)

	return lTmpl
}

// 创建一个Template
func (self *TTemplate) New(name, htmlstream string, langmap map[string]string) (*TTemplate, error) {
	//    cwd := os.Getenv("CWD")
	// 创建词法分析器
	xlexer := &lexer{
		input:      htmlstream,
		leftdelim:  LEFT_DELIM,
		rightdelim: RIGHT_DELIM,
		pos:        0,
		curline:    1}

	// 创建分析器
	xparser := &Parser{
		template: self,
		parent:   nil,
		lex:      xlexer,
		silo:     []interface{}{},
	}

	xtmpl := &TTemplate{Name: name,
		parser:  xparser,
		langMap: langmap,
	} //创建一个新的Template

	return xtmpl, nil
}

// 分析文件或数据流
func (self *TTemplate) Parse(aSrc string /*, langmap map[string]string*/) (*TTemplate, error) {
	/*
			//var stream []byte
			var lName, lFileName, lFileDir string
			//var err error

			// aSrc 可以是文件或数据流
			_, e := os.Stat(aSrc) // 返回FileInfo 只返回当有该文件时,其他出错

			//log.Println(e)
			//log.Println(d, e, aSrc, !d.IsDir(), cc, aa, path.IsAbs(cc))
			if len(aSrc) < 255 && e == nil { //e==nil 代表有该文件.所以是文件
				lFileDir, lFileName = filepath.Split(aSrc)

				stream, err := ioutil.ReadFile(aSrc) // 读取文件
				if err != nil {
					return nil, err
				}

				aSrc = string(stream)

			} else {
				name = "Unknow" //获得文件名 " index.htm"
				//log.Println("IsStream")
			}

		//创建一个新的Template
		//xtmpl := &Template{Name: name, Dir: lFileDir, langMap: langmap}
		self.Name = lFileName
		self.Dir = lFileDir
		self.langMap = langmap
	*/
	if _, err := self.parse(aSrc); err != nil {
		return nil, err
	}

	return self, nil
}

// 解析Template 调用New 创建 调用parser的parse实现解析
func (self *TTemplate) parse(aText string) (*TTemplate, error) {
	// 判断是否创建好

	if self.parser == nil {
		// 创建词法分析器
		xlex := &lexer{
			input:      aText,
			leftdelim:  LEFT_DELIM,
			rightdelim: RIGHT_DELIM,
			pos:        0,
			curline:    1}

		// 创建分析器
		self.parser = &Parser{template: self,
			parent: nil,
			lex:    xlex,
			silo:   []interface{}{}}
	} else {
		self.parser.lex.input = aText
	}

	//self.parser.lex.stream = stream
	err := self.parser.parse()
	if err != nil {
		return self, err
	}

	return self, nil
}

// 主要渲染
func (self *TTemplate) Render(context ...interface{}) string {
	var buf bytes.Buffer
	var contextChain []interface{}

	// 循环获得模板变量数据值
	for _, c := range context {
		val := reflect.ValueOf(c)
		contextChain = append(contextChain, val)
	}

	self.parser.renderTemplate(contextChain, &buf) // 返回渲染好的Html到Buf
	return buf.String()
}

func (self *TTemplate) RenderToWriter(data map[string]interface{}, w io.Writer) error {
	// 合并模板变量
	err := self.template.Execute(w, data) // 绘制模板
	if err != nil {
		return err
	}

	return nil
}

func Escape(content string) string {
	return template.HTMLEscapeString(content)
}

func FromString(template string) (res_html *TTemplateSet) {
	return defaultTemplateSet
}

func FromFile(template string) (res_html *TTemplateSet) {

	return defaultTemplateSet
}

/*
//添加数据集供模板渲染，value==nil时删除
func Context(key string, value interface{}) {

}
*/

// 添加方法,fn为Nil时为删除
func Func(name string, fn interface{}) {

}

func Str2html(raw string) template.HTML {
	return template.HTML(raw)

	//fmt.Println("rawsssssssssss:", raw)
	//return template.HTMLEscaper(raw)
}

func SetStaticPath(path string) {
	TEMPLATES_ROOT = path
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
