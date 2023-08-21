package python_test

import (
	"testing"
	"testing/fstest"

	"github.com/nedpals/errgoengine/error_templates/python"
	testutils "github.com/nedpals/errgoengine/error_templates/test_utils"
)

var cases = testutils.TestCases{
	"name_error": {
		Input: `Traceback (most recent call last):
  File "name_error.py", line 1, in <module>
    print(test)
          ^^^^
NameError: name 'test' is not defined`,
		ExpectedLanguage: "Python",
		ExpectedTemplate: "NameError",
		Expected:         "Your program tried to access the 'test' variable which was not found on your program.",
		Files: fstest.MapFS{
			"name_error.py": {
				Data: []byte(`print(test)`),
			},
		},
	},
	"zero_division_error": {
		Input: `Traceback (most recent call last):
  File "zero_division_error.py", line 1, in <module>
    print(1 / 0)
          ~~^~~
ZeroDivisionError: division by zero`,
		ExpectedLanguage: "Python",
		ExpectedTemplate: "ZeroDivisionError",
		Expected:         "TODO",
		Files: fstest.MapFS{
			"zero_division_error.py": {
				Data: []byte(`print(1 / 0)`),
			},
		},
	},
}

func TestPythonErrorTemplates(t *testing.T) {
	cases.Execute(t, python.LoadErrorTemplates)
}
