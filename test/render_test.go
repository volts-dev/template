package test

import (
	"fmt"
	"testing"

	"github.com/volts-dev/template"
	//"github.com/volts-dev/volts/server"
	"github.com/volts-dev/utils"
)

var AppFilePath = utils.AppPath()

func TestRender(t *testing.T) {
	BaseEngine := template.NewBaseEngine()

	//lHtml, err := BaseEngine.RenderTemplate(nil, AppFilePath+"\\src\\base.html", nil)
	lHtml, err := BaseEngine.RenderTemplate(nil, "./src/base.html", nil)
	fmt.Println("Html", lHtml, err)

}
