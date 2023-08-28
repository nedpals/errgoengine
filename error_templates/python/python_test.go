package python_test

import (
	"testing"

	"github.com/nedpals/errgoengine/error_templates/python"
	testutils "github.com/nedpals/errgoengine/error_templates/test_utils"
)

func TestPythonErrorTemplates(t *testing.T) {
	testutils.SetupTest(t, testutils.SetupTestConfig{
		DirName:        "python",
		TemplateLoader: python.LoadErrorTemplates,
	}).Execute(t)
}
