package template

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

type (
	TBaseContext map[string]interface{}

	TBaseEngine struct {
		template_set *TTemplateSet
	}
)

func init() {
	// 注册模板引擎
	RegisterEngine("base", &TBaseEngine{})
}

func NewBaseEngine() *TBaseEngine {
	return &TBaseEngine{}
}

func (self *TBaseEngine) TemplateSet() *TTemplateSet {
	return self.template_set
}

// TODO
func (self *TBaseEngine) RenderTemplate(template_set *TTemplateSet, loader ILoader, template string, context map[string]interface{}) (res_doc string, res_err error) {
	// 可以是文件或数据流
	//_, res_err = os.Stat(template) // 返回FileInfo 只返回当有该文件时,其他出错
	//if res_err != nil {
	//		return
	//	}

	//log.Println(e)
	//log.Println(d, e, aSrc, !d.IsDir(), cc, aa, path.IsAbs(cc))
	var lFileDir string
	//if (len(template) < 255 || strings.HasPrefix(template, "<")) && res_err == nil { //e==nil 代表有该文件.所以是文件
	if strings.HasPrefix(template, "<") || strings.HasPrefix(template, "{") {
		//name = "Unknow" //获得文件名 " index.htm"

		//log.Println("IsStream")
		res_doc = template
	} else {
		lFileDir, _ = filepath.Split(template) // [文件名]
		//log.Println("lFileDir")
		lStream, err := ioutil.ReadFile(template) // 读取文件
		if err != nil {
			return "", err
		}
		//log.Println("lFileDir2", lStream)
		res_doc = string(lStream)
	}

	//TODO 改为工厂模式
	lTemplate := NewTemplate("")
	lTemplate.Dir = lFileDir
	lTemplate.Parse(res_doc) //, langmap) // 分析文件或数据流
	return lTemplate.Render(context), nil

}
