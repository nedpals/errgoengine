package errgoengine_test

import (
	"context"
	"testing"

	lib "github.com/nedpals/errgoengine"
	testutils "github.com/nedpals/errgoengine/test_utils"
)

type TestAnalyzer struct{}

func (TestAnalyzer) FallbackSymbol() lib.Symbol {
	return nil
}

func (TestAnalyzer) FindSymbol(string) lib.Symbol { return nil }

func (TestAnalyzer) AnalyzeNode(context.Context, lib.SyntaxNode) lib.Symbol {
	return nil
}

func (TestAnalyzer) AnalyzeImport(lib.ImportParams) lib.ResolvedImport {
	return lib.ResolvedImport{}
}

var testLanguage = &lib.Language{
	Name:              "TestLang",
	FilePatterns:      []string{".test"},
	StackTracePattern: `\sin (?P<symbol>\S+) at (?P<path>\S+):(?P<position>\d+)`,
	LocationConverter: func(ctx lib.LocationConverterContext) lib.Location {
		return lib.Location{
			DocumentPath: ctx.Path,
			StartPos:     lib.Position{0, 0, 0},
			EndPos:       lib.Position{0, 0, 0},
		}
	},
	AnalyzerFactory: func(cd *lib.ContextData) lib.LanguageAnalyzer {
		return TestAnalyzer{}
	},
}

func emptyExplainFn(cd *lib.ContextData, gen *lib.ExplainGenerator) {}
func emptyBugFixFn(cd *lib.ContextData, gen *lib.BugFixGenerator)   {}

func setupTemplate(template lib.ErrorTemplate) (*lib.CompiledErrorTemplate, error) {
	errorTemplates := lib.ErrorTemplates{}
	return errorTemplates.Add(testLanguage, template)
}

func TestErrorTemplate(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		tmp, err := setupTemplate(lib.ErrorTemplate{
			Name:           "SampleError",
			Pattern:        "This is a sample error",
			OnGenExplainFn: emptyExplainFn,
			OnGenBugFixFn:  emptyBugFixFn,
		})

		if err != nil {
			t.Fatal(err)
		}

		testutils.Equals(t, tmp.Name, "SampleError")
		testutils.Equals(t, tmp.Language, testLanguage)
		testutils.Equals(t, tmp.Pattern.String(), `(?m)^This is a sample error(?P<stacktrace>(?:.|\s)*)$`)
		testutils.Equals(t, tmp.StackTraceRegex().String(), `(?m)\sin (?P<symbol>\S+) at (?P<path>\S+):(?P<position>\d+)`)
		testutils.ExpectNil(t, tmp.StackTracePattern)
	})

	t.Run("With custom stack trace", func(t *testing.T) {
		tmp, err := setupTemplate(lib.ErrorTemplate{
			Name:              "SampleError2",
			Pattern:           "This is a sample error with stack trace",
			OnGenExplainFn:    emptyExplainFn,
			OnGenBugFixFn:     emptyBugFixFn,
			StackTracePattern: `(?P<symbol>\S+):(?P<path>\S+):(?P<position>\d+)`,
		})

		if err != nil {
			t.Fatal(err)
		}

		testutils.Equals(t, tmp.Name, "SampleError2")
		testutils.Equals(t, tmp.Language, testLanguage)
		testutils.Equals(t, tmp.Pattern.String(), `(?m)^This is a sample error with stack trace(?P<stacktrace>(?:.|\s)*)$`)
		testutils.Equals(t, tmp.StackTraceRegex().String(), `(?P<symbol>\S+):(?P<path>\S+):(?P<position>\d+)`)
	})

	t.Run("With custom error pattern", func(t *testing.T) {
		tmp, err := setupTemplate(lib.ErrorTemplate{
			Name:           "SampleError3",
			Pattern:        lib.CustomErrorPattern("Stack trace in middle $stacktracetest"),
			OnGenExplainFn: emptyExplainFn,
			OnGenBugFixFn:  emptyBugFixFn,
		})

		if err != nil {
			t.Fatal(err)
		}

		testutils.Equals(t, tmp.Name, "SampleError3")
		testutils.Equals(t, tmp.Language, testLanguage)
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

		testutils.Equals(t, tmp.StackTraceRegex().String(), "(?m)"+testLanguage.StackTracePattern)
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
