package python_test

import (
	"testing"

	python "github.com/nedpals/errgoengine/languages/python"
	ltutils "github.com/nedpals/errgoengine/languages/test_utils"
)

func TestPython(t *testing.T) {
	cases := ltutils.TestCases{
		ltutils.TestCase{
			Name:     "Simple",
			FileName: "simple.py",
			Input: `
def main():
	a = 1
			`,
			Expected: `
(tree [0,0 | 0]-[1,6 | 18]
	(function  [0,0 | 0]-[1,6 | 18]
		(tree [0,0 | 0]-[1,6 | 18]
			(assignment a [1,1 | 13]-[1,2 | 14]))))
`,
		},
	}

	cases.Execute(t, python.Language)
}
