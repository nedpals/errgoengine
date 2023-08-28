package java_test

import (
	"testing"

	"github.com/nedpals/errgoengine/error_templates/java"
	testutils "github.com/nedpals/errgoengine/error_templates/test_utils"
)

func TestJavaErrorTemplates(t *testing.T) {
	testutils.SetupTest(t, testutils.SetupTestConfig{
		DirName:        "java",
		TemplateLoader: java.LoadErrorTemplates,
	}).Execute(t)
}
