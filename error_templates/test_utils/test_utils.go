package testutils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	lib "github.com/nedpals/errgoengine"
)

type TestCase struct {
	Input            string
	ExpectedLanguage string
	ExpectedTemplate string
	Expected         string
	Files            fstest.MapFS
}

type TestCases map[string]TestCase

func (cases TestCases) Execute(t *testing.T, loadTemplate func(*lib.ErrorTemplates)) {
	// createdTemplates := fs.Glob(os., pattern)
	wd, _ := os.Getwd()
	name := filepath.Base(wd)
	implTemplates, err := fs.Glob(os.DirFS(wd), "*.go")
	if err != nil {
		t.Fatal(err)
	}

	engine := lib.New("")
	loadTemplate(&engine.ErrorTemplates)
	engine.ErrorTemplates.CompileAll()

	for _, templatePath := range implTemplates {
		if strings.HasSuffix(fmt.Sprintf("%s.go", name), templatePath) || strings.HasSuffix(fmt.Sprintf("%s_test.go", name), templatePath) {
			continue
		}

		templateName := strings.TrimSuffix(filepath.Base(templatePath), filepath.Ext(templatePath))
		t.Run(templateName, func(t *testing.T) {
			tCase, exists := cases[templateName]
			if !exists {
				t.Fatal("Test case not implemented")
			}

			engine.FS = tCase.Files
			trimmedInput := strings.TrimSpace(tCase.Input)
			template, data, err := engine.Analyze(trimmedInput)
			if err != nil {
				t.Fatal(err)
			} else if template.Name != tCase.ExpectedTemplate || template.Language.Name != tCase.ExpectedLanguage {
				t.Fatalf(
					"\nExpected: (Language: %s, Template: %s)\nGot:      (Language: %s, Template: %s)",
					tCase.ExpectedLanguage, tCase.ExpectedTemplate,
					template.Language.Name, template.Name,
				)
			}

			trimmedExpected := strings.TrimSpace(tCase.Expected)
			output := engine.Translate(template, data)
			if output != trimmedExpected {
				t.Errorf("\nExpected: %s\nGot:      %s", tCase.Expected, output)
			}
		})
	}
}
