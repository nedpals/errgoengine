package errgoengine_test

import (
	"strings"
	"testing"

	lib "github.com/nedpals/errgoengine"
	sitter "github.com/smacker/go-tree-sitter"
)

func TestOutputGenerator(t *testing.T) {
	parser := sitter.NewParser()
	doc, err := lib.ParseDocument("program.test", strings.NewReader("a = xyz\nb = 123"), parser, lib.TestLanguage, nil)
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
` + "```" + ``

		if output != expected {
			t.Errorf("exp %s, got %s", expected, output)
		}
	})

	t.Run("With sections", func(t *testing.T) {
		defer gen.Reset()

		bugFix := lib.NewBugFixGenerator(doc)
		explain := lib.NewExplainGeneratorForError("NameError")

		// create a fake name error explanation
		explain.Add("The variable you are trying to use is not defined. In this case, the variable `xyz` is not defined.")

		// add a section
		explain.CreateSection("More info").
			Add("This error is usually caused by a typo or a missing variable definition.")

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
## More info
This error is usually caused by a typo or a missing variable definition.
## Steps to fix
### Define the variable ` + "`xyz`" + ` before using it.
In line 1, replace ` + "`xyz`" + ` with ` + "`\"test\"`" + `.
` + "```diff" + `
- a = xyz
+ a = "test"
b = 123
` + "```" + ``
		if output != expected {
			t.Errorf("exp %s, got %s", expected, output)
		}
	})

	t.Run("With multiple suggestions", func(t *testing.T) {
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

		bugFix.Add("Define the variable `xyz` before using it.", func(s *lib.BugFixSuggestion) {
			s.AddStep("In line 1, declare a new variable named `xyz`").
				AddFix(lib.FixSuggestion{
					NewText: "xyz = \"test\"\n",
					StartPosition: lib.Position{
						Line:   0,
						Column: 0,
					},
					EndPosition: lib.Position{
						Line:   0,
						Column: 0,
					},
				})
		})

		// generate the output
		output := gen.Generate(explain, bugFix)

		// check if the output is correct
		expected := `# NameError
The variable you are trying to use is not defined. In this case, the variable ` + "`xyz`" + ` is not defined.
## Steps to fix
### 1. Define the variable ` + "`xyz`" + ` before using it.
In line 1, replace ` + "`xyz`" + ` with ` + "`\"test\"`" + `.
` + "```diff" + `
- a = xyz
+ a = "test"
b = 123
` + "```" + `

### 2. Define the variable ` + "`xyz`" + ` before using it.
In line 1, declare a new variable named ` + "`xyz`" + `.
` + "```diff" + `
- a = xyz
+ xyz = "test"
+ a = xyz
b = 123
` + "```" + ``

		if output != expected {
			t.Errorf("exp %s, got %s", expected, output)
		}
	})

	t.Run("With fix description", func(t *testing.T) {
		defer gen.Reset()

		bugFix := lib.NewBugFixGenerator(doc)
		explain := lib.NewExplainGeneratorForError("NameError")

		// create a fake name error explanation
		explain.Add("The variable you are trying to use is not defined. In this case, the variable `xyz` is not defined.")

		// create a fake bug fix suggestion
		bugFix.Add("Define the variable `xyz` before using it.", func(s *lib.BugFixSuggestion) {
			s.AddStep("In line 1, replace `xyz` with `\"test\"`.").
				AddFix(lib.FixSuggestion{
					Description: "This is a test description.",
					NewText:     "\"test\"",
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
` + "```" + `
This is a test description.`

		if output != expected {
			t.Errorf("exp %s, got %s", expected, output)
		}
	})

	t.Run("With GenAfterExplain", func(t *testing.T) {
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

		// add a code snippet that points to the error
		gen.GenAfterExplain = func(gen *lib.OutputGenerator) {
			startLineNr := 0
			startLines := doc.LinesAt(startLineNr, startLineNr+1)
			endLines := doc.LinesAt(startLineNr+1, startLineNr+2)

			gen.Writeln("```")
			gen.WriteLines(startLines...)

			for i := 0; i < 4; i++ {
				if startLines[len(startLines)-1][i] == '\t' {
					gen.Builder.WriteString("    ")
				} else {
					gen.Builder.WriteByte(' ')
				}
			}

			for i := 0; i < 3; i++ {
				gen.Builder.WriteByte('^')
			}

			gen.Break()
			gen.WriteLines(endLines...)
			gen.Writeln("```")
		}

		// generate the output
		output := gen.Generate(explain, bugFix)

		// check if the output is correct
		expected := `# NameError
The variable you are trying to use is not defined. In this case, the variable ` + "`xyz`" + ` is not defined.
` + "```" + `
a = xyz
b = 123
    ^^^
b = 123
` + "```" + `
## Steps to fix
### Define the variable ` + "`xyz`" + ` before using it.
In line 1, replace ` + "`xyz`" + ` with ` + "`\"test\"`" + `.
` + "```diff" + `
- a = xyz
+ a = "test"
b = 123
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
