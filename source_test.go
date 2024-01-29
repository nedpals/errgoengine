package errgoengine

import (
	"context"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

var testLanguage = &Language{
	Name:              "Python",
	FilePatterns:      []string{".py"},
	SitterLanguage:    python.GetLanguage(),
	StackTracePattern: `\s+File "(?P<path>\S+)", line (?P<position>\d+), in (?P<symbol>\S+)`,
	ErrorPattern:      `Traceback \(most recent call last\):$stacktrace$message`,
	AnalyzerFactory: func(cd *ContextData) LanguageAnalyzer {
		return &testAnalyzer{cd}
	},
	SymbolsToCapture: `
(expression_statement
	(assignment
		left: (identifier) @assignment.name
		right: (identifier) @assignment.content) @assignment)
`,
}

type testAnalyzer struct {
	*ContextData
}

func (an *testAnalyzer) FallbackSymbol() Symbol {
	return Builtin("any")
}

func (an *testAnalyzer) FindSymbol(name string) Symbol {
	return nil
}

func (an *testAnalyzer) AnalyzeNode(_ context.Context, n SyntaxNode) Symbol {
	// TODO:
	return Builtin("void")
}

func (an *testAnalyzer) AnalyzeImport(params ImportParams) ResolvedImport {
	// TODO:

	return ResolvedImport{
		Path: "",
	}
}

func TestParseDocument(t *testing.T) {
	parser := sitter.NewParser()

	doc, err := ParseDocument("test", strings.NewReader(`hello = 1`), parser, testLanguage, nil)
	if err != nil {
		t.Error(err)
	}

	if doc.Contents != "hello = 1" {
		t.Errorf("Expected contents to be \"hello = 1\", got %q", doc.Contents)
	}
}

func TestEditableDocument(t *testing.T) {
	parser := sitter.NewParser()

	doc, err := ParseDocument("test", strings.NewReader(`hello = 1`), parser, testLanguage, nil)
	if err != nil {
		t.Error(err)
	} else if doc.TotalLines() < 1 {
		t.Errorf("Expected document to have at least 1 line, got %d", doc.TotalLines())
	}

	// Add
	t.Run("EditableDocument.Add", func(t *testing.T) {
		editableDoc := doc.Editable()

		editableDoc.Apply(Changeset{
			NewText: "_world",
			StartPos: Position{
				Line:   0,
				Column: 5,
				Index:  5,
			},
			EndPos: Position{
				Line:   0,
				Column: 5,
				Index:  5,
			},
		})

		if editableDoc.String() != "hello_world = 1" {
			t.Errorf("Expected contents to be \"hello_world = 1\", got %q", editableDoc.String())
		}
	})

	t.Run("EditableDocument.AddMultipleLines", func(t *testing.T) {
		editableDoc := doc.Editable()

		addedLines := []string{
			"println('hello world!')",
			"world = 2",
		}

		editableDoc.Apply(Changeset{
			NewText: strings.Join(addedLines, "\n") + "\n",
			StartPos: Position{
				Line:   0,
				Column: 0,
				Index:  5,
			},
			EndPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
		})

		if len(editableDoc.changesets) != 3 {
			t.Errorf("Expected changesets to be 3, got %d", len(editableDoc.changesets))
		}

		if len(editableDoc.modifiedLines) != 3 {
			t.Errorf("Expected cached lines to be 3, got %d", len(editableDoc.modifiedLines))
		}

		for idx, line := range addedLines {
			if editableDoc.modifiedLines[idx] != line {
				t.Errorf("Expected contents to be %q on line %d, got %q", line, idx, editableDoc.modifiedLines[idx])
			}
		}

		exp := strings.Join(addedLines, "\n") + "\n" + "hello = 1"
		if editableDoc.String() != exp {
			t.Errorf("Expected contents to be %q, got %q", exp, editableDoc.String())
		}
	})

	t.Run("EditableDocument.AddMultipleLinesDouble", func(t *testing.T) {
		editableDoc := doc.Editable()

		addedLines := []string{
			"println('hello world!')",
			"world = 2",
		}

		editableDoc.Apply(Changeset{
			NewText: strings.Join(addedLines, "\n") + "\n\n",
			StartPos: Position{
				Line:   0,
				Column: 0,
				Index:  5,
			},
			EndPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
		})

		if len(editableDoc.changesets) != 4 {
			t.Errorf("Expected changesets to be 4, got %d", len(editableDoc.changesets))
		}

		if len(editableDoc.modifiedLines) != 4 {
			t.Errorf("Expected cached lines to be 4, got %d", len(editableDoc.modifiedLines))
		}

		for idx, line := range addedLines {
			if editableDoc.modifiedLines[idx] != line {
				t.Errorf("Expected contents to be %q on line %d, got %q", line, idx, editableDoc.modifiedLines[idx])
			}
		}

		exp := strings.Join(addedLines, "\n") + "\n\n" + "hello = 1"
		if editableDoc.String() != exp {
			t.Errorf("Expected contents to be %q, got %q", exp, editableDoc.String())
		}
	})

	t.Run("EditableDocument.AddAndExtendLine", func(t *testing.T) {
		editableDoc := doc.Editable()
		addedLines := []string{
			"println('hello world!')",
			"foo_",
		}

		editableDoc.Apply(Changeset{
			NewText: strings.Join(addedLines, "\n"),
			StartPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
			EndPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
		})

		if len(editableDoc.changesets) != 2 {
			t.Errorf("Expected changesets to be 2, got %d", len(editableDoc.changesets))
		}

		if len(editableDoc.modifiedLines) != 2 {
			t.Errorf("Expected cached lines to be 2, got %d", len(editableDoc.modifiedLines))
		}

		for idx, line := range addedLines {
			if idx == len(addedLines)-1 {
				if editableDoc.modifiedLines[idx] != "foo_hello = 1" {
					t.Errorf("Expected contents to be %q on line %d, got %q", "foo_hello = 1", idx, editableDoc.modifiedLines[idx])
				}
			} else if editableDoc.modifiedLines[idx] != line {
				t.Errorf("Expected contents to be %q on line %d, got %q", line, idx, editableDoc.modifiedLines[idx])
			}
		}

		exp := strings.Join(addedLines, "\n") + "hello = 1"
		if editableDoc.String() != exp {
			t.Errorf("Expected contents to be %q, got %q", exp, editableDoc.String())
		}
	})

	t.Run("EditableDocument.AddMultipleLines2", func(t *testing.T) {
		editableDoc := doc.Editable()

		addedLines := []string{
			"println('hello world!')",
			"{",
			"}",
		}

		editableDoc.Apply(Changeset{
			NewText: strings.Join(addedLines, "\n") + "\n",
			StartPos: Position{
				Line:   0,
				Column: 0,
				Index:  5,
			},
			EndPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
		})

		if len(editableDoc.changesets) != 4 {
			t.Errorf("Expected changesets to be 4, got %d", len(editableDoc.changesets))
		}

		if len(editableDoc.modifiedLines) != 4 {
			t.Errorf("Expected cached lines to be 4, got %d", len(editableDoc.modifiedLines))
		}

		for idx, line := range addedLines {
			if editableDoc.modifiedLines[idx] != line {
				t.Errorf("Expected contents to be %q on line %d, got %q", line, idx, editableDoc.modifiedLines[idx])
			}
		}

		exp := strings.Join(addedLines, "\n") + "\nhello = 1"
		if editableDoc.String() != exp {
			t.Errorf("Expected contents to be %q, got %q", exp, editableDoc.String())
		}
	})

	t.Run("EditableDocument.AddMultipleLinesAfter", func(t *testing.T) {
		editableDoc := doc.Editable()

		addedLines := []string{
			"println('hello world!')",
			"world = 2",
		}

		editableDoc.Apply(Changeset{
			NewText: "\n" + strings.Join(addedLines, "\n"),
			StartPos: Position{
				Line:   0,
				Column: 9,
				Index:  9,
			},
			EndPos: Position{
				Line:   0,
				Column: 9,
				Index:  9,
			},
		})

		exp := "hello = 1\n" + strings.Join(addedLines, "\n")
		if editableDoc.String() != exp {
			t.Errorf("Expected contents to be %q, got %q", exp, editableDoc.String())
		}
	})

	t.Run("EditableDocument.AddMultipleLinesAfterDouble", func(t *testing.T) {
		editableDoc := doc.Editable()
		editableDoc.Apply(Changeset{
			NewText: "\n\n" + "println('hello world!')",
			StartPos: Position{
				Line:   0,
				Column: 9,
				Index:  9,
			},
			EndPos: Position{
				Line:   0,
				Column: 9,
				Index:  9,
			},
		})

		exp := "hello = 1\n\n" + "println('hello world!')"
		if editableDoc.String() != exp {
			t.Errorf("Expected contents to be %q, got %q", exp, editableDoc.String())
		}
	})

	t.Run("EditableDocument.AddMultipleLinesMiddle", func(t *testing.T) {
		editableDoc := doc.Editable()

		addedLines := []string{
			"println('hello world!')",
			"world = 2",
		}

		editableDoc.Apply(Changeset{
			NewText: "\n" + strings.Join(addedLines, "\n") + "\n",
			StartPos: Position{
				Line:   0,
				Column: 2,
				Index:  2,
			},
			EndPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
		})

		exp := "he\n" + strings.Join(addedLines, "\n") + "\nllo = 1"
		if editableDoc.String() != exp {
			t.Errorf("Expected contents to be %q, got %q", exp, editableDoc.String())
		}
	})

	// Replace
	t.Run("EditableDocument.Replace", func(t *testing.T) {
		editableDoc := doc.Editable()

		editableDoc.Apply(Changeset{
			NewText: "world",
			StartPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
			EndPos: Position{
				Line:   0,
				Column: 5,
				Index:  5,
			},
		})

		if editableDoc.String() != "world = 1" {
			t.Errorf("Expected contents to be \"world = 1\", got %q", editableDoc.String())
		}
	})

	t.Run("EditableDocument.ReplaceWithPadding", func(t *testing.T) {
		doc, err := ParseDocument("test", strings.NewReader(`        hello = 1`), parser, testLanguage, nil)
		if err != nil {
			t.Error(err)
		} else if doc.TotalLines() < 1 {
			t.Errorf("Expected document to have at least 1 line, got %d", doc.TotalLines())
		}

		editableDoc := doc.Editable()

		editableDoc.Apply(Changeset{
			NewText: "world",
			StartPos: Position{
				Line:   0,
				Column: 8,
				Index:  8,
			},
			EndPos: Position{
				Line:   0,
				Column: 13,
				Index:  13,
			},
		})

		if editableDoc.String() != "        world = 1" {
			t.Errorf("Expected contents to be \"        world = 1\", got %q", editableDoc.String())
		}
	})

	t.Run("EditableDocument.ReplaceMultipleLines", func(t *testing.T) {
		editableDoc := doc.Editable()

		editableDoc.Apply(Changeset{
			NewText: "wwo = 3\nfoo",
			StartPos: Position{
				Line:   0,
				Column: 2,
				Index:  3,
			},
			EndPos: Position{
				Line:   0,
				Column: 5,
				Index:  5,
			},
		})

		// hello = 1
		if editableDoc.String() != "hewwo = 3\nfoo = 1" {
			t.Errorf("Expected contents to be \"hewwo = 3\\nfoo = 1\", got %q", editableDoc.String())
		}
	})

	// Remove
	t.Run("EditableDocument.Remove", func(t *testing.T) {
		editableDoc := doc.Editable()

		editableDoc.Apply(Changeset{
			StartPos: Position{
				Line:   0,
				Column: 1,
				Index:  1,
			},
			EndPos: Position{
				Line:   0,
				Column: 3,
				Index:  3,
			},
		})

		if editableDoc.String() != "hlo = 1" {
			t.Errorf("Expected contents to be \"hlo = 1\", got %q", editableDoc.String())
		}
	})

	t.Run("EditableDocument.RemoveMultipleLines", func(t *testing.T) {
		editableDoc := doc.Editable()

		editableDoc.Apply(Changeset{
			NewText: "a = 1\nb = 2\n",
			StartPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
			EndPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
		})

		if editableDoc.String() != "a = 1\nb = 2\nhello = 1" {
			t.Errorf("Expected contents to be \"a = 1\\nb = 2\\nhello = 1\", got %q", editableDoc.String())
		}

		editableDoc.Apply(Changeset{
			StartPos: Position{
				Line:   1,
				Column: 0,
				Index:  6,
			},
			EndPos: Position{
				Line:   2,
				Column: 9,
				Index:  22,
			},
		})

		if editableDoc.String() != "a = 1" {
			t.Errorf("Expected contents to be \"a = 1\", got %q", editableDoc.String())
		}
	})

	t.Run("EditableDocument.RemoveMultipleLinesMiddle", func(t *testing.T) {
		editableDoc := doc.Editable()

		editableDoc.Apply(Changeset{
			NewText: "a = 1\nb = 2\nc = 3\n",
			StartPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
			EndPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
		})

		if editableDoc.String() != "a = 1\nb = 2\nc = 3\nhello = 1" {
			t.Errorf("Expected contents to be \"a = 1\\nb = 2\\nc = 3\\nhello = 1\", got %q", editableDoc.String())
		}

		editableDoc.Apply(Changeset{
			StartPos: Position{
				Line:   1,
				Column: 0,
			},
			EndPos: Position{
				Line:   2,
				Column: 5,
			},
		})

		if editableDoc.String() != "a = 1\nhello = 1" {
			t.Errorf("Expected contents to be \"a = 1\", got %q", editableDoc.String())
		}
	})

	t.Run("EditableDocument.RemoveMultipleLines2", func(t *testing.T) {
		editableDoc := doc.Editable()

		editableDoc.Apply(Changeset{
			NewText: "a = 1\nb = 2\n",
			StartPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
			EndPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
		})

		if editableDoc.String() != "a = 1\nb = 2\nhello = 1" {
			t.Errorf("Expected contents to be \"a = 1\\nb = 2\\nhello = 1\", got %q", editableDoc.String())
		}

		editableDoc.Apply(Changeset{
			StartPos: Position{
				Line:   1,
				Column: 0,
				Index:  6,
			},
			EndPos: Position{
				Line:   2,
				Column: 1,
				Index:  9,
			},
		})

		if editableDoc.String() != "a = 1\nello = 1" {
			t.Errorf("Expected contents to be \"a = 1\\nello = 1\", got %q", editableDoc.String())
		}
	})

	t.Run("EditableDocument.WrapWithBlock", func(t *testing.T) {
		editableDoc := doc.Editable()

		editableDoc.Apply(Changeset{
			NewText: "inner text\n{\n\t",
			StartPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
			EndPos: Position{
				Line:   0,
				Column: 0,
				Index:  0,
			},
		})

		editableDoc.Apply(Changeset{
			NewText: "\n}",
			StartPos: Position{
				Line:   2,
				Column: 10,
				Index:  0,
			},
			EndPos: Position{
				Line:   2,
				Column: 10,
				Index:  0,
			},
		})

		editableDoc.Apply(Changeset{
			NewText: "if {\n\t\t",
			StartPos: Position{
				Line:   2,
				Column: 1,
				Index:  0,
			},
			EndPos: Position{
				Line:   2,
				Column: 1,
				Index:  0,
			},
		})

		if editableDoc.String() != "inner text\n{\n\tif {\n\t\thello = 1\n}" {
			t.Errorf("Expected contents to be \"inner text\\n{\\n\\tif {\\n\\t\\thello = 1\\n}\", got %q", editableDoc.String())
		}

		editableDoc.Apply(Changeset{
			NewText: "\t\n}",
			StartPos: Position{
				Line:   3,
				Column: 11,
				Index:  0,
			},
			EndPos: Position{
				Line:   3,
				Column: 11,
				Index:  0,
			},
		})

		if editableDoc.String() != "inner text\n{\n\tif {\n\t\thello = 1\t\n}\n}" {
			t.Errorf("Expected contents to be \"inner text\\n{\\n\\tif {\\n\\t\\thello = 1\\t\\n}\\n}\", got %q", editableDoc.String())
		}
	})
}
