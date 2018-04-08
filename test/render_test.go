package test

import (
	"fmt"
	"testing"
	"vectors/utils"
	"vectors/web"
	"vectors/web/template"
)

var AppFilePath = utils.AppPath()

func TestRender(t *testing.T) {
	BaseEngine := template.NewBaseEngine()

	//lHtml, err := BaseEngine.RenderTemplate(nil, AppFilePath+"\\src\\base.html", nil)
	lHtml, err := BaseEngine.RenderTemplate(nil, "./src/base.html", nil)
	fmt.Println("Html", lHtml, err)

}
