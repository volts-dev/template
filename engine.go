package template

import (
	//	"fmt"
	"strings"

	"github.com/volts-dev/logger"
)

type (
	//TODO 改成Reader 提供模板加载的接口
	ILoader interface {
		// TODO ReadSource() 提供模板加载的接口
		ReadTemplate(name string) string
		// TODO WriteSource()/ 提供静态文件保存方法
		WriteTemplate(template map[string]interface{}) string

		// 读取模板里资源文件 name 为文件路径名
		ReadAsset(xmlid, path, inc, typ string) []*TAssetFile
		WriteAsset(template map[string]interface{})
	}

	IEngine interface {
		RenderTemplate(template_set *TTemplateSet, loader ILoader, template string, context map[string]interface{}) (string, error)
		TemplateSet() *TTemplateSet
	}
)

var (
	template_engines = map[string]IEngine{"html": NewBaseEngine()} // 注册的Writer类型函数接口
)

func RegisterEngine(name string, engine IEngine) {
	name = strings.ToLower(name)
	if engine == nil {
		panic("template engine: Register provide is nil")
	}
	if _, dup := template_engines[name]; dup {
		panic("template engine: Register called twice for provider " + name)
	}
	template_engines[name] = engine
}

func NewEngine(name string) (engine IEngine) {
	var has bool
	if engine, has = template_engines[name]; !has || engine == nil {
		//logger.Logger.Err("%s engine not available!", name)

		if engine, has = template_engines["html"]; !has || engine == nil {
			logger.Logger.Err("%s engine not available!", name)
			return nil //fmt.Errorf("%s engine not available!", name)
		}
	}

	return
}
