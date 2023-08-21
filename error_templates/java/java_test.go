package java_test

import (
	"testing"
	"testing/fstest"

	"github.com/nedpals/errgoengine/error_templates/java"
	testutils "github.com/nedpals/errgoengine/error_templates/test_utils"
)

var cases = testutils.TestCases{
	"null_pointer_exception": {
		Input: `Exception in thread "main" java.lang.NullPointerException
    at ShouldBeNull.main(ShouldBeNull.java:4)`,
		Expected:         `Your program tried to print the value of "toUpperCase" method from "test" which is a null.`,
		ExpectedLanguage: "Java",
		ExpectedTemplate: "NullPointerException",
		Files: fstest.MapFS{
			"ShouldBeNull.java": {
				Data: []byte(`public class ShouldBeNull {
	public static void main(String args[]) {
		String test = null;
		System.out.println(test.toUpperCase());
	}
}`),
			},
		},
	},
	"arithmetic_exception": {
		Input: `Exception in thread "main" java.lang.ArithmeticException: / by zero
        at Arith.main(Arith.java:3)`,
		ExpectedLanguage: "Java",
		ExpectedTemplate: "ArithmeticException",
		Expected:         "One of your variables initialized a double value by dividing a number to zero",
		Files: fstest.MapFS{
			"Arith.java": {
				Data: []byte(`public class Arith {
	public static void main(String[] args) {
		double out = 3 / 0;
		System.out.println(out);
	}
}`),
			},
		},
	},
	"array_index_out_of_bounds_exception": {
		Input: `Exception in thread "main" java.lang.ArrayIndexOutOfBoundsException: Index 5 out of bounds for length 4
        at OOB.main(OOB.java:4)`,
		ExpectedLanguage: "Java",
		ExpectedTemplate: "ArrayIndexOutOfBoundsException",
		Expected:         "Your program attempted to access an element in index 5 on an array that has only 4 items",
		Files: fstest.MapFS{
			"OOB.java": {
				Data: []byte(`public class Arith {
	public static void main(String[] args) {
		double out = 3 / 0;
		System.out.println(out);
	}
}`),
			},
		},
	},
	"array_required_type_error": {
		Input: `NotArray.java:4: error: array required, but int found
	        int value = number[0];
	                          ^
	1 error`,
		ExpectedLanguage: "Java",
		ExpectedTemplate: "ArrayRequiredTypeError",
		Expected:         "Your program attempted to access an element in index 5 on an array that has only 4 items",
		Files: fstest.MapFS{
			"OOB.java": {
				Data: []byte(`public class NotArray {
		public static void main(String[] args) {
			int number = 5;
			int value = number[0];
		}
	}`),
			},
		},
	},
	"parse_end_of_file_error": {
		Input: `EOF.java:4: error: reached end of file while parsing
    }
     ^
1 error`,
		ExpectedLanguage: "Java",
		ExpectedTemplate: "ParseEndOfFileError",
		Expected:         "Your program attempted to access an element in index 5 on an array that has only 4 items",
		Files: fstest.MapFS{
			"EOF.java": {
				Data: []byte(`public class EOF {
public static void main(String[] args) {
	System.out.println("This is a sample program.");
}
`),
			},
		},
	},
	"public_class_filename_mismatch_error": {
		Input: `Wrong.java:1: error: class Right is public, should be declared in a file named Right.java
		public class Right {
			   ^
		1 error`,
		ExpectedLanguage: "Java",
		ExpectedTemplate: "PublicClassFilenameMismatchError",
		Expected:         "Your program attempted to access an element in index 5 on an array that has only 4 items",
		Files: fstest.MapFS{
			"Wrong.java": {
				Data: []byte(`public class Right {
	public static void main(String[] args) {

	}
}`),
			},
		},
	},
	"unreachable_statement_error": {
		Input: `Unreachable.java:5: error: unreachable statement
        System.out.println("c");
        ^
1 error`,
		ExpectedLanguage: "Java",
		ExpectedTemplate: "UnreachableStatementError",
		Expected:         "Your program attempted to access an element in index 5 on an array that has only 4 items",
		Files: fstest.MapFS{
			"Unreachable.java": {
				Data: []byte(`public class Unreachable {
	public static void main(String[] args) {
		System.out.println("b");
		return;
		System.out.println("c");
	}
}`),
			},
		},
	},
}

func TestJavaErrorTemplates(t *testing.T) {
	cases.Execute(t, java.LoadErrorTemplates)
}
