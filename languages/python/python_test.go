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
class Square:
	side = 1

	def __init__(self, side):
		self.side = side

	def area(self):
		return self.side * self.side

	def perimeter(self):
		return 4 * self.side

def main():
	a = 1
			`,
			Expected: `
(tree [0,0 | 0]-[13,6 | 185]
	(class Square Square [0,0 | 0]-[10,22 | 165]
		(tree [1,1 | 15]-[10,22 | 165]
			(assignment int side [1,1 | 15]-[1,5 | 19])
			(function any __init__ [3,1 | 26]-[4,18 | 70]
				(tree [3,1 | 26]-[4,18 | 70]
					(variable any self [3,13 | 38]-[3,25 | 50])
					(variable any side [3,13 | 38]-[3,25 | 50])))
			(function any area [6,1 | 73]-[7,30 | 119]
				(tree [6,1 | 73]-[7,30 | 119]
					(variable any self [6,9 | 81]-[6,15 | 87])))
			(function any perimeter [9,1 | 122]-[10,22 | 165]
				(tree [9,1 | 122]-[10,22 | 165]
					(variable any self [9,14 | 135]-[9,20 | 141])))))
	(function any main [12,0 | 167]-[13,6 | 185]
		(tree [12,0 | 167]-[13,6 | 185]
			(assignment int a [13,1 | 180]-[13,2 | 181]))))
`,
		},
	}

	cases.Execute(t, python.Language)
}
