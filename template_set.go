package template

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/volts-dev/cacher"
	"github.com/volts-dev/cacher/memory"
)

type (

	// 管理模板对象  变量以及缓存等
	TTemplateSet struct { //TemplateSet
		name    string
		FuncMap map[string]interface{} //储存[模板函数]
		VarMap  map[string]interface{} //储存[模板变量]

		Cacher map[string]memory.ListCache

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

func Default() *TTemplateSet {
	return defaultTemplateSet
}

func New() *TTemplateSet {
	tmpl := &TTemplateSet{
		Cacher:  make(map[string]memory.ListCache),
		FuncMap: make(map[string]interface{}),
		VarMap:  make(map[string]interface{}),
	}
	tmpl.AddFuncs(TemplateFuncs)
	return tmpl
}

func (self *TTemplateSet) AddVar(vars map[string]interface{}) {
	for key, val := range vars {
		self.VarMap[key] = val
	}
}
func (self *TTemplateSet) AddFuncs(funcs map[string]interface{}) {
	for key, val := range funcs {
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
		block := c.Front()
		if tmpl, ok := block.Value.(*template.Template); ok {
			lTmpl = tmpl
			///log.Println("Template in cache is vaild", tmpl)
		} else {
			///log.Println("Template in cache is invaild : %s")
			return
		}
	} else {
		tmpl := NewTemplate(lName)
		tmpl.DirPath = lFileDir
		tmpl.Parse(aHtmlSrc) //, langmap) // 分析文件或数据流

		// 使用原生态Parse 扩展处理好的HTML
		lTmpl = template.New(lName).Funcs(self.FuncMap) //## 考虑是否有必要添加Funcs 到lTmpl.Execute(w, data)前添加模板函数

		_, err := lTmpl.Parse(tmpl.Render(data)) // 分析
		if err != nil {
			log.Errf("X-Error:tmpl.Parse %s", err)
		}

		//log.Println("RenderToResponse:", tmpl.Render(data))
		// 为新的模板新建缓存系统
		if self.Cacher[lName] == nil {
			self.Cacher[lName] = memory.New()
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
		self.Cacher[lName].Set(&cacher.CacheBlock{Key: lName, Value: lTmpl})
	}

}

func (self *TTemplateSet) GetTemplate(templateName string) *template.Template {
	// Name of template
	name := self.MD5(templateName)
	// 是否存在该 Template 的缓存且不为空
	if c, ok := self.Cacher[name]; ok && c.Len() > 0 {
		block := c.Front()
		if tmpl, ok := block.Value.(*template.Template); ok {
			return tmpl
		}
	}

	return nil
}

func (self *TTemplateSet) RenderToString(template_name string, data map[string]interface{}, engine_args ...interface{}) (html string, err error) {
	var buf bytes.Buffer
	err = self.RenderToWriter(template_name, data, &buf, engine_args...)
	return buf.String(), err
}

func (self *TTemplateSet) RenderToBytes(template_name string, data map[string]interface{}, engine_args ...interface{}) (html []byte, err error) {
	var buf bytes.Buffer
	err = self.RenderToWriter(template_name, data, &buf, engine_args...)
	return buf.Bytes(), err
}

func (self *TTemplateSet) RenderToWriter(template_name string, data map[string]interface{}, w io.Writer, engine_args ...interface{}) error {
	var (
		lName       string
		lTmpl       *template.Template
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

	engine, err := NewEngine(lEngineName)
	if err != nil {
		return err
	}

	// Name of template
	lName = self.MD5(template_name)

	// 是否存在该 Template 的缓存且不为空
	if c, ok := self.Cacher[lName]; ok && c.Len() > 0 {
		block := c.Front()
		if tmpl, ok := block.Value.(*template.Template); ok {
			lTmpl = tmpl
		}
	}

	if lTmpl == nil {
		// TODO FIX 这里可能无法缓存data需要修改
		lTmplStr, err := engine.RenderTemplate(loader, template_name, data)
		if err != nil {
			return err
		}
		// 使用原生态Parse 扩展处理好的HTML
		lTmpl = template.New(lName).Funcs(self.FuncMap) //## 考虑是否有必要添加Funcs 到lTmpl.Execute(w, data)前添加模板函数

		_, err = lTmpl.Parse(lTmplStr) // 分析
		if err != nil {
			return err
		}

		// 为新的模板新建缓存系统
		if self.Cacher[lName] == nil {
			self.Cacher[lName] = memory.New()
		}
	}

	// 合并模板变量
	//data = utils.MergeMaps(self.VarMap, data)
	err = lTmpl.Execute(w, data) // 绘制模板
	if err != nil {
		return err
	}

	// 保留缓存
	if self.Cacheable {
		self.Cacher[lName].Set(&cacher.CacheBlock{Key: lName, Value: lTmpl})
	}

	return nil
}
