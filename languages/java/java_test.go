package java_test

import (
	"testing"

	java "github.com/nedpals/errgoengine/languages/java"
	ltutils "github.com/nedpals/errgoengine/languages/test_utils"
)

func TestJava(t *testing.T) {
	cases := ltutils.TestCases{
		ltutils.TestCase{
			Name:     "Simple",
			FileName: "Test.java",
			Input: `
public class Test {
	public static void main(String[] args) {
		int a = 1;
		double c = 0.0d;
	}

	public int add(int a, int b) {
		return a + b;
	}
}
			`,
			Expected: `
(tree [0,0 | 0]-[9,1 | 150]
	(class Test [0,0 | 0]-[9,1 | 150]
		(tree [0,18 | 18]-[9,1 | 150]
			(function main [1,1 | 21]-[4,2 | 96]
				(tree [1,1 | 21]-[4,2 | 96]
					(variable String[] args [1,25 | 45]-[1,38 | 58])
					(variable int a [2,6 | 68]-[2,11 | 73])
					(variable double c [3,9 | 84]-[3,17 | 92])))
			(function add [6,1 | 99]-[8,2 | 148]
				(tree [6,1 | 99]-[8,2 | 148]
					(variable int a [6,16 | 114]-[6,21 | 119])
					(variable int b [6,23 | 121]-[6,28 | 126]))))))
			`,
		},
	}

	cases.Execute(t, java.Language)
}
