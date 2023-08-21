package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/nedpals/errgoengine/error_templates"
	"github.com/nedpals/errgoengine/lib"
)

var errorTemplates = lib.ErrorTemplates{}

func init() {
	error_templates.LoadErrorTemplates(&errorTemplates)
	errorTemplates.CompileAll()
}

func main() {
	var errMsg string
	wd, _ := os.Getwd()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if len(errMsg) != 0 {
			errMsg += "\n"
		}

		errMsg += scanner.Text()
	}

	if len(errMsg) == 0 {
		os.Exit(1)
	}

	engine := lib.New(wd)
	engine.ErrorTemplates = errorTemplates
	template, data, err := engine.Analyze(errMsg)

	if err != nil {
		panic(err)
	}

	fmt.Println(engine.Translate(template, data))
}
