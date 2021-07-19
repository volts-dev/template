package template

import (
	"os"
	"path"
)

const (
	LEFT_DELIM  = "{%"
	RIGHT_DELIM = "%}"
)

var (
	//	Tpl *Template // 模板库
	ROOT, _ = os.Getwd()
	//	MEDIA_ROOT     = path.Join(ROOT, "/static")    // 静态文件物理路径
	TEMPLATES_ROOT = path.Join(ROOT, "/template") // 模板路径
)
