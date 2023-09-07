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
(tree [0,0]-[0,0]
	(function main [0,0]
		(tree [1,1]-[1,6]
			(variable a [1,1]))))
`,
		},
	}

	cases.Execute(t, python.Language)
}
