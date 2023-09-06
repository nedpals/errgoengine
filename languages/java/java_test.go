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
(tree [0,0]-[0,0]
	(class Test [0,0]
		(tree [0,18]-[1,1]
			(function main [1,1]
				(tree [1,40]-[2,2]
					(variable a [2,2]))))))
			`,
		},
	}

	cases.Execute(t, java.Language)
}
