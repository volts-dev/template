package template

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/VectorsOrigin/cacher"
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
		set     *TTemplateSet
		Name    string
		Dir     string
		parser  *Parser
		langMap map[string]string

		level int
	}

	// 管理模板对象  变量以及缓存等
	TTemplateSet struct { //TemplateSet
		name    string
		FuncMap map[string]interface{} //储存[模板函数]
		VarMap  map[string]interface{} //储存[模板变量]

		Cacher map[string]cache.ICacher

		// Sandbox features
		// - Disallow access to specific tags and/or filters (using BanTag() and BanFilter())
		//
		// For efficiency reasons you can ban tags/filters only *before* you have
		// added your first template to the set (restrictions are statically checked).
		// After you added one, it's not possible anymore (for your personal security).
		firstTemplateCreated bool
		bannedTags           map[string]bool
		bannedFilters        map[string]bool
		Cacheable            bool // 可缓存
		lock                 sync.Mutex
	}
)

// Version string
const Version = "beta"

var (
	//tplcache cache.ICache
	TemplateFuncs = map[string]interface{}{
		"trans": func(aText string) template.HTML {
			return template.HTML(aText)
			// 在Router里实现了！
		},
		"Str2html": Str2html,
	}
	TemplateVars = map[string]interface{}{} //储存[模板变量]

	TemplateSet = NewTemplateSet()
)

func init() {
	//TemplateFuncs = make(map[string]interface{})

	/*
		// 添加默认[模板函数]
		AddTmplFuncs(map[string]interface{}{
			"Str2html": Str2html,
		},
		)
	*/
	// 注册新缓存
	//cache.Register("tplcache", cache.NewMemoryCache())
	//tplcache, _ = cache.NewCache("tplcache", `{"interval":1}`)
	//tplcache.Active(false)
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

func (self *TTemplateSet) AddVar(aVarMap map[string]interface{}) {
	for key, val := range aVarMap {
		self.VarMap[key] = val
	}
}
func (self *TTemplateSet) AddFuncs(aFuncMap map[string]interface{}) {
	for key, val := range aFuncMap {
		self.FuncMap[key] = val
	}
}

// Name of template
func (self *TTemplateSet) MD5(name string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(name)))
}

// X+Org渲染
func (self *TTemplateSet) __Render(engine IEngine, aHtmlSrc string, w http.ResponseWriter /* langmap map[string]string, */, data interface{}) {
	//var stream []byte
	var (
		lName, lFileDir string
		//lFileName string
		lTmpl *template.Template
	)
	//var err error

	// 可以是文件或数据流
	_, e := os.Stat(aHtmlSrc) // 返回FileInfo 只返回当有该文件时,其他出错
	//log.Println(e)
	//log.Println(d, e, aSrc, !d.IsDir(), cc, aa, path.IsAbs(cc))
	if len(aHtmlSrc) < 255 && e == nil { //e==nil 代表有该文件.所以是文件
		lFileDir, _ = filepath.Split(aHtmlSrc) // [文件名]

		lStream, err := ioutil.ReadFile(aHtmlSrc) // 读取文件
		if err != nil {
			return // nil, err
		}

		aHtmlSrc = string(lStream)
	} else {
		//name = "Unknow" //获得文件名 " index.htm"

		//log.Println("IsStream")
	}
	// Name of template
	lName = fmt.Sprintf("%x", md5.Sum([]byte(aHtmlSrc)))
	///log.Println("Template", lName)
	// 是否存在该 Template 的缓存且不为空
	if c, ok := self.Cacher[lName]; ok && c.Len() > 0 {
		if tmpl, ok := c.Front().(*template.Template); ok {
			lTmpl = tmpl
			///log.Println("Template in cache is vaild", tmpl)
		} else {
			///log.Println("Template in cache is invaild : %s")
			return
		}
	} else {
		tmpl := NewTemplate(lName)
		tmpl.Dir = lFileDir
		tmpl.Parse(aHtmlSrc) //, langmap) // 分析文件或数据流

		// 使用原生态Parse 扩展处理好的HTML
		lTmpl = template.New(lName).Funcs(self.FuncMap) //## 考虑是否有必要添加Funcs 到lTmpl.Execute(w, data)前添加模板函数

		_, err := lTmpl.Parse(tmpl.Render(data)) // 分析
		if err != nil {
			log.Println("X-Error:tmpl.Parse %s", err)
		}

		//log.Println("RenderToResponse:", tmpl.Render(data))
		// 为新的模板新建缓存系统
		if self.Cacher[lName] == nil {
			self.Cacher[lName] = cache.NewMemoryCache()
		}
		//log.Println("New Template in cache is vaildable", tmpl)
	}

	// 合并模板变量
	//data = utils.MergeMaps(self.VarMap, data)
	err := lTmpl.Execute(w, data) // 绘制模板
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 保留缓存
	if self.Cacheable {
		self.Cacher[lName].Push(lTmpl, 43200)
	}

}

func (self *TTemplateSet) RenderToString(template_name string, data map[string]interface{}, engine_args ...interface{}) (res_html string, res_err error) {
	var buf bytes.Buffer
	res_err = self.RenderToWriter(template_name, data, &buf, engine_args...)
	return buf.String(), res_err
}

func (self *TTemplateSet) RenderToBytes(template_name string, data map[string]interface{}, engine_args ...interface{}) (res_html []byte, res_err error) {
	var buf bytes.Buffer
	res_err = self.RenderToWriter(template_name, data, &buf, engine_args...)
	return buf.Bytes(), res_err
}

func (self *TTemplateSet) RenderToWriter(template_name string, data map[string]interface{}, w io.Writer /* langmap map[string]string, */, engine_args ...interface{}) (res_err error) {
	//var stream []byte
	var (
		lName string
		//lFileName string
		lTmpl       *template.Template
		engine      IEngine
		lEngineName string
		loader      ILoader
	)
	// 拆分引擎参数为 name,loader
	for idex, arg := range engine_args {
		if idex == 0 { // engine_name
			if a, allowed := arg.(string); allowed && a != "" {
				lEngineName = a
			}
		}

		if idex == 1 {
			if ldr, allowed := arg.(ILoader); allowed && ldr != nil {
				loader = ldr
			}
		}
	}

	engine = NewEngine(lEngineName)

	// Name of template
	lName = self.MD5(template_name)

	///log.Println("Template", lName)
	// 是否存在该 Template 的缓存且不为空
	if c, ok := self.Cacher[lName]; ok && c.Len() > 0 {
		if tmpl, ok := c.Front().(*template.Template); ok {
			lTmpl = tmpl
			log.Println("Template in cache is vaild", tmpl)
		}
	}

	if lTmpl == nil {
		lTmplStr := ""
		lTmplStr, res_err = engine.RenderTemplate(self, loader, template_name, data)
		if res_err != nil {
			return
		}
		// 使用原生态Parse 扩展处理好的HTML
		lTmpl = template.New(lName).Funcs(self.FuncMap) //## 考虑是否有必要添加Funcs 到lTmpl.Execute(w, data)前添加模板函数

		//log.Println("htmlhtml:", lTmplStr)
		_, res_err = lTmpl.Parse(lTmplStr) // 分析
		if res_err != nil {
			//log.Println("X-Error:tmpl.Parse %s", res_err)
			//http.Error(w, res_err.Error(), http.StatusInternalServerError)
			//fmt.Fprintln(w, error)
			return res_err
		}

		//log.Println("RenderToResponse:", template_name, lTmplStr)
		// 为新的模板新建缓存系统
		if self.Cacher[lName] == nil {
			self.Cacher[lName] = cache.NewMemoryCache()
		}
		//log.Println("New Template in cache is vaildable", tmpl)
	}

	// 合并模板变量
	//data = utils.MergeMaps(self.VarMap, data)
	res_err = lTmpl.Execute(w, data) // 绘制模板
	if res_err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return res_err
	}

	// 保留缓存
	if self.Cacheable {
		self.Cacher[lName].Push(lTmpl, 43200)
	}

	return
}

func NewTemplateSet() *TTemplateSet {
	lHtml := &TTemplateSet{
		Cacher:  make(map[string]cache.ICacher),
		FuncMap: make(map[string]interface{}),
		VarMap:  make(map[string]interface{}),
	}
	lHtml.AddFuncs(TemplateFuncs)
	return lHtml
}

func Escape(content string) string {
	return template.HTMLEscapeString(content)
}

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
		silo:     []interface{}{}}

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
	_, err := self.parse(aSrc)
	if err != nil {
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
	//log.Println(buf.String())
	return buf.String()
}

func FromString(template string) (res_html *TTemplateSet) {
	return TemplateSet
}

func FromFile(template string) (res_html *TTemplateSet) {

	return TemplateSet
}

/*
//添加数据集供模板渲染，value==nil时删除
func Context(key string, value interface{}) {

}
*/

// 添加方法,fn为Nil时为删除
func Func(name string, fn interface{}) {

}
