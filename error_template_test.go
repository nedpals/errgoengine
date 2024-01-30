package errgoengine_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	lib "github.com/nedpals/errgoengine"
	testutils "github.com/nedpals/errgoengine/test_utils"
)

func emptyExplainFn(cd *lib.ContextData, gen *lib.ExplainGenerator) {}
func emptyBugFixFn(cd *lib.ContextData, gen *lib.BugFixGenerator)   {}

func setupTemplate(template lib.ErrorTemplate) (*lib.CompiledErrorTemplate, error) {
	errorTemplates := lib.ErrorTemplates{}
	return errorTemplates.Add(lib.TestLanguage, template)
}

func TestErrorTemplate(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		tmp, err := setupTemplate(lib.ErrorTemplate{
			Name:           "ErrorA",
			Pattern:        "This is a sample error",
			OnGenExplainFn: emptyExplainFn,
			OnGenBugFixFn:  emptyBugFixFn,
		})

		if err != nil {
			t.Fatal(err)
		}

		testutils.Equals(t, tmp.Name, "ErrorA")
		testutils.Equals(t, tmp.Language, lib.TestLanguage)
		testutils.Equals(t, tmp.Pattern.String(), `(?m)^This is a sample error(?P<stacktrace>(?:.|\s)*)$`)
		testutils.Equals(t, tmp.StackTraceRegex().String(), `(?m)\sin (?P<symbol>\S+) at (?P<path>\S+):(?P<position>\d+)`)
		testutils.ExpectNil(t, tmp.StackTracePattern)
	})

	t.Run("With custom stack trace", func(t *testing.T) {
		tmp, err := setupTemplate(lib.ErrorTemplate{
			Name:              "ErrorB",
			Pattern:           "This is a sample error with stack trace",
			OnGenExplainFn:    emptyExplainFn,
			OnGenBugFixFn:     emptyBugFixFn,
			StackTracePattern: `(?P<symbol>\S+):(?P<path>\S+):(?P<position>\d+)`,
		})

		if err != nil {
			t.Fatal(err)
		}

		testutils.Equals(t, tmp.Name, "ErrorB")
		testutils.Equals(t, tmp.Language, lib.TestLanguage)
		testutils.Equals(t, tmp.Pattern.String(), `(?m)^This is a sample error with stack trace(?P<stacktrace>(?:.|\s)*)$`)
		testutils.Equals(t, tmp.StackTraceRegex().String(), `(?P<symbol>\S+):(?P<path>\S+):(?P<position>\d+)`)
	})

	t.Run("With custom error pattern", func(t *testing.T) {
		tmp, err := setupTemplate(lib.ErrorTemplate{
			Name:           "ErrorC",
			Pattern:        lib.CustomErrorPattern("Stack trace in middle $stacktracetest"),
			OnGenExplainFn: emptyExplainFn,
			OnGenBugFixFn:  emptyBugFixFn,
		})

		if err != nil {
			t.Fatal(err)
		}

		testutils.Equals(t, tmp.Name, "ErrorC")
		testutils.Equals(t, tmp.Language, lib.TestLanguage)
		testutils.Equals(t, tmp.Pattern.String(), `(?m)^Stack trace in middle (?P<stacktrace>(?:.|\s)*)test$`)
		testutils.ExpectNil(t, tmp.StackTracePattern)
	})
}

func TestStackTraceRegex(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		tmp, err := setupTemplate(lib.ErrorTemplate{
			Name:           "A",
			Pattern:        "AA",
			OnGenExplainFn: emptyExplainFn,
			OnGenBugFixFn:  emptyBugFixFn,
		})
		if err != nil {
			t.Fatal(err)
		}

		testutils.Equals(t, tmp.StackTraceRegex().String(), "(?m)"+lib.TestLanguage.StackTracePattern)
	})

	t.Run("With custom stack trace", func(t *testing.T) {
		tmp, err := setupTemplate(lib.ErrorTemplate{
			Name:              "B",
			Pattern:           "BB",
			OnGenExplainFn:    emptyExplainFn,
			OnGenBugFixFn:     emptyBugFixFn,
			StackTracePattern: `(?P<symbol>\S+):(?P<path>\S+):(?P<position>\d+)`,
		})

		if err != nil {
			t.Fatal(err)
		}

		testutils.Equals(t, tmp.StackTraceRegex().String(), `(?P<symbol>\S+):(?P<path>\S+):(?P<position>\d+)`)
	})

	t.Run("Should Error", func(t *testing.T) {
		defer func() {
			if err, ok := recover().(string); ok {
				testutils.Equals(t, err, "expected stacktrace pattern got compiled, got nil regex instead")
			}
		}()

		errorTemplates := lib.ErrorTemplates{}
		errLang := &lib.Language{Name: "Err", StackTracePattern: "aa", AnalyzerFactory: func(cd *lib.ContextData) lib.LanguageAnalyzer { return nil }}
		errLang.Compile()

		lib.SetTemplateStackTraceRegex(errLang, nil)

		tmp, err := errorTemplates.Add(errLang, lib.ErrorTemplate{
			Name:           "B",
			Pattern:        "BB",
			OnGenExplainFn: emptyExplainFn,
			OnGenBugFixFn:  emptyBugFixFn,
		})
		if err != nil {
			t.Fatal(err)
		}

		func() {
			_ = tmp.StackTraceRegex()
		}()

		t.Fatal("expected panic, got successful execution instead")
	})

}

func TestExtractVariables(t *testing.T) {
	tmp, err := setupTemplate(lib.ErrorTemplate{
		Name:           "WithVarError",
		Pattern:        "invalid input '(?P<input>.*)'",
		OnGenExplainFn: emptyExplainFn,
		OnGenBugFixFn:  emptyBugFixFn,
	})
	if err != nil {
		t.Fatal(err)
	}

	tmp2, err := setupTemplate(lib.ErrorTemplate{
		Name:           "WithoutVarError",
		Pattern:        "invalid input '.*'",
		OnGenExplainFn: emptyExplainFn,
		OnGenBugFixFn:  emptyBugFixFn,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Simple", func(t *testing.T) {
		input := "invalid input '123abc'\nin main at /home/user/main.py:123\nin main at /home/user/main.py:1"
		if !tmp.Match(input) {
			t.Fatalf("expected template to match input, got false instead")
		}

		variables := tmp.ExtractVariables(input)
		exp := map[string]string{
			"stacktrace": "\nin main at /home/user/main.py:123\nin main at /home/user/main.py:1",
			"input":      "123abc",
		}

		fmt.Printf("%q\n", variables)
		if !reflect.DeepEqual(variables, exp) {
			t.Fatalf("expected %v, got %v", exp, variables)
		}
	})

	t.Run("No stack trace", func(t *testing.T) {
		input := "invalid input 'wxyz88@'"
		if !tmp.Match(input) {
			t.Fatalf("expected template to match input, got false instead")
		}

		variables := tmp.ExtractVariables(input)
		exp := map[string]string{
			"input":      "wxyz88@",
			"stacktrace": "",
		}

		fmt.Printf("%q\n", variables)
		if !reflect.DeepEqual(variables, exp) {
			t.Fatalf("expected %v, got %v", exp, variables)
		}
	})

	t.Run("No variables", func(t *testing.T) {
		input := "invalid input '123abc'\nin main at /home/user/main.py:123\nin main at /home/user/main.py:1"
		if !tmp2.Match(input) {
			t.Fatalf("expected template to match input, got false instead")
		}

		variables := tmp2.ExtractVariables(input)
		exp := map[string]string{
			"stacktrace": "\nin main at /home/user/main.py:123\nin main at /home/user/main.py:1",
		}

		fmt.Printf("%q\n", variables)
		if !reflect.DeepEqual(variables, exp) {
			t.Fatalf("expected %v, got %v", exp, variables)
		}
	})

	t.Run("No variables + no stack trace", func(t *testing.T) {
		input := "invalid input '123abc'"
		if !tmp2.Match(input) {
			t.Fatalf("expected template to match input, got false instead")
		}

		variables := tmp2.ExtractVariables(input)
		exp := map[string]string{
			"stacktrace": "",
		}

		fmt.Printf("%q\n", variables)
		if !reflect.DeepEqual(variables, exp) {
			t.Fatalf("expected %v, got %v", exp, variables)
		}
	})
}

func TestExtractStackTrace(t *testing.T) {
	t.Run("Extract stack trace", func(t *testing.T) {
		tmp, err := setupTemplate(lib.ErrorTemplate{
			Name:           "A",
			Pattern:        "invalid input '(?P<input>.*)'",
			OnGenExplainFn: emptyExplainFn,
			OnGenBugFixFn:  emptyBugFixFn,
		})
		if err != nil {
			t.Fatal(err)
		}

		cd := lib.NewContextData(lib.NewEmptyStore(), "/home/user")
		cd.Variables = map[string]string{
			"stacktrace": "\nin main at /home/user/main.py:123\nin main at /home/user/main.py:1",
		}

		stackTrace := tmp.ExtractStackTrace(cd)
		exp := lib.TraceStack{
			lib.StackTraceEntry{
				SymbolName: "main",
				Location: lib.Location{
					DocumentPath: "/home/user/main.py",
					StartPos: lib.Position{
						Line:   123,
						Column: 0,
						Index:  0,
					},
					EndPos: lib.Position{
						Line:   123,
						Column: 0,
						Index:  0,
					},
				},
			},
			lib.StackTraceEntry{
				SymbolName: "main",
				Location: lib.Location{
					DocumentPath: "/home/user/main.py",
					StartPos: lib.Position{
						Line:   1,
						Column: 0,
						Index:  0,
					},
					EndPos: lib.Position{
						Line:   1,
						Column: 0,
						Index:  0,
					},
				},
			},
		}

		if !reflect.DeepEqual(stackTrace, exp) {
			t.Fatalf("expected %v, got %v", exp, stackTrace)
		}
	})
}

func TestErrorTemplates(t *testing.T) {
	errorTemplates := lib.ErrorTemplates{}
	tmp := errorTemplates.MustAdd(lib.TestLanguage, lib.ErrorTemplate{
		Name:           "ErrorA",
		Pattern:        `This is a sample error\n`,
		OnGenExplainFn: emptyExplainFn,
		OnGenBugFixFn:  emptyBugFixFn,
	})

	tmp2 := errorTemplates.MustAdd(lib.TestLanguage, lib.ErrorTemplate{
		Name:           "ErrorB",
		Pattern:        `Another exmaple error\n`,
		OnGenExplainFn: emptyExplainFn,
		OnGenBugFixFn:  emptyBugFixFn,
	})

	fmt.Println(tmp.Pattern.String())
	fmt.Println(tmp2.Pattern.String())

	t.Run("Simple", func(t *testing.T) {
		inputs := []string{
			"This is a sample error",
			"Another exmaple error",
		}

		expected := []*lib.CompiledErrorTemplate{
			tmp,
			tmp2,
		}

		for i, input := range inputs {
			matched := errorTemplates.Match(input + "\n" + lib.TestLanguage.StackTracePattern)

			if !reflect.DeepEqual(matched, expected[i]) {
				t.Fatalf("expected %s, got %s", expected[i].Name, matched.Name)
			}
		}
	})

	t.Run("SimpleReverse", func(t *testing.T) {
		inputs := []string{
			"Another exmaple error",
			"This is a sample error",
		}

		expected := []*lib.CompiledErrorTemplate{
			tmp2,
			tmp,
		}

		for i, input := range inputs {
			matched := errorTemplates.Match(input + "\n" + lib.TestLanguage.StackTracePattern)

			if !reflect.DeepEqual(matched, expected[i]) {
				t.Fatalf("expected %s, got %s", expected[i].Name, matched.Name)
			}
		}
	})

	t.Run("Should be nil", func(t *testing.T) {
		inputs := []string{
			"This is a sample errorz\n",
			"AAnother exmaple error\n",
			"Another eaaxmaple error\n" + lib.TestLanguage.StackTracePattern,
			"This is a sample erroar\n" + lib.TestLanguage.StackTracePattern,
		}

		for _, input := range inputs {
			matched := errorTemplates.Match(input)

			if matched != nil {
				t.Fatalf("expected nil, got %s", matched.Name)
			}
		}
	})

	t.Run("Stacked", func(t *testing.T) {
		inputs := []string{
			"This is a sample error",
			"Another exmaple error",
		}

		input := strings.Join(inputs, "\nin main at /home/user/main.py:1\n\n")
		matched := errorTemplates.Match(input)

		if !reflect.DeepEqual(matched, tmp) {
			t.Fatalf("expected %s, got %s", tmp.Name, matched.Name)
		}
	})

	t.Run("StackedReverse", func(t *testing.T) {
		// reverse
		inputs := []string{
			"Another exmaple error",
			"This is a sample error",
		}

		input := strings.Join(inputs, "\nin main at /home/user/main.py:1\n\n")
		matched := errorTemplates.Match(input)

		if !reflect.DeepEqual(matched, tmp2) {
			t.Fatalf("expected %s, got %s", tmp2.Name, matched.Name)
		}
	})
}
