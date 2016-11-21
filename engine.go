package template

import (
	//	"fmt"
	"strings"
	"webgo/logger"
)

type (
	//TODO 提供模板加载的接口
	ILoader interface {
		ReadTemplate(name string) string
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
