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
	}
}
			`,
			Expected: `
(tree [0,0 | 0]-[4,1 | 79]
	(class Test [0,0 | 0]-[4,1 | 79]
		(tree [0,18 | 18]-[4,1 | 79]
			(function main [1,1 | 21]-[3,2 | 77]
				(tree [1,1 | 21]-[3,2 | 77]
					(variable a [2,6 | 68]-[2,11 | 73]))))))
			`,
		},
	}

	cases.Execute(t, java.Language)
}
