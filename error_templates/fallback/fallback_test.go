package error_templates_test

import (
	"testing"

	lib "github.com/nedpals/errgoengine"
	testutils "github.com/nedpals/errgoengine/error_templates/test_utils"
)

func TestFallbackErrorTemplate(t *testing.T) {
	testutils.SetupTest(t, testutils.SetupTestConfig{
		DirName: "fallback",
		TemplateLoader: func(et *lib.ErrorTemplates) {
			(*et)[lib.TemplateKey("", lib.FallbackErrorTemplate.Name)] = lib.FallbackErrorTemplate
		},
	}).Execute(t)
}
