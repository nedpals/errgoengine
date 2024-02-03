package errgoengine_test

import (
	"strings"
	"testing"

	lib "github.com/nedpals/errgoengine"
	sitter "github.com/smacker/go-tree-sitter"
)

func TestOutputGenerator(t *testing.T) {
	parser := sitter.NewParser()
	doc, err := lib.ParseDocument("program.test", strings.NewReader("a = xyz\nb = 123\nxyz = \"test\""), parser, lib.TestLanguage, nil)
	if err != nil {
		t.Fatal(err)
	}

	gen := &lib.OutputGenerator{}

	t.Run("Simple", func(t *testing.T) {
		defer gen.Reset()

		bugFix := lib.NewBugFixGenerator(doc)
		explain := lib.NewExplainGeneratorForError("NameError")

		// create a fake name error explanation
		explain.Add("The variable you are trying to use is not defined. In this case, the variable `xyz` is not defined.")

		// create a fake bug fix suggestion
		bugFix.Add("Define the variable `xyz` before using it.", func(s *lib.BugFixSuggestion) {
			s.AddStep("In line 1, replace `xyz` with `\"test\"`.").
				AddFix(lib.FixSuggestion{
					NewText: "\"test\"",
					StartPosition: lib.Position{
						Line:   0,
						Column: 4,
					},
					EndPosition: lib.Position{
						Line:   0,
						Column: 7,
					},
				})
		})

		// generate the output
		output := gen.Generate(explain, bugFix)

		// check if the output is correct
		expected := `# NameError
The variable you are trying to use is not defined. In this case, the variable ` + "`xyz`" + ` is not defined.
## Steps to fix
### Define the variable ` + "`xyz`" + ` before using it.
In line 1, replace ` + "`xyz`" + ` with ` + "`\"test\"`" + `.
` + "```diff" + `
- a = xyz
+ a = "test"
b = 123
xyz = "test"
` + "```" + ``

		if output != expected {
			t.Errorf("exp %s, got %s", expected, output)
		}
	})

	t.Run("Empty explanation", func(t *testing.T) {
		defer gen.Reset()

		bugFix := lib.NewBugFixGenerator(doc)
		explain := lib.NewExplainGeneratorForError("NameError")

		// generate bug fix
		bugFix.Add("Define the variable `xyz` before using it.", func(s *lib.BugFixSuggestion) {
			s.AddStep("In line 1, replace `xyz` with `\"test\"`.").
				AddFix(lib.FixSuggestion{
					NewText: "\"test\"",
					StartPosition: lib.Position{
						Line:   0,
						Column: 4,
					},
					EndPosition: lib.Position{
						Line:   0,
						Column: 7,
					},
				})
		})

		// generate the output
		output := gen.Generate(explain, bugFix)

		// check if the output is correct
		expected := `# NameError
No explanation found for this error.
## Steps to fix
### Define the variable ` + "`xyz`" + ` before using it.
In line 1, replace ` + "`xyz`" + ` with ` + "`\"test\"`" + `.
` + "```diff" + `
- a = xyz
+ a = "test"
b = 123
xyz = "test"
` + "```" + ``
		if output != expected {
			t.Errorf("exp %s, got %s", expected, output)
		}
	})

	t.Run("Empty bug fixes", func(t *testing.T) {
		defer gen.Reset()

		bugFix := lib.NewBugFixGenerator(doc)
		explain := lib.NewExplainGeneratorForError("NameError")

		// create a fake name error explanation
		explain.Add("The variable you are trying to use is not defined. In this case, the variable `xyz` is not defined.")

		// generate the output
		output := gen.Generate(explain, bugFix)

		// check if the output is correct
		expected := `# NameError
The variable you are trying to use is not defined. In this case, the variable ` + "`xyz`" + ` is not defined.
## Steps to fix
No bug fixes found for this error.`
		if output != expected {
			t.Errorf("exp %s, got %s", expected, output)
		}
	})

	t.Run("Empty explanation + bug fixes", func(t *testing.T) {
		defer gen.Reset()

		bugFix := lib.NewBugFixGenerator(doc)
		explain := lib.NewExplainGeneratorForError("NameError")

		// generate the output
		output := gen.Generate(explain, bugFix)

		// check if the output is correct
		expected := `# NameError
No explanation found for this error.
## Steps to fix
No bug fixes found for this error.`
		if output != expected {
			t.Errorf("exp %s, got %s", expected, output)
		}
	})
}
