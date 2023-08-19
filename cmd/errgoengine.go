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

	fmt.Println(lib.Analyze(errorTemplates, wd, errMsg))
}
