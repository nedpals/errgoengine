package errgoengine

import (
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
		return &pyAnalyzer{cd}
	},
	SymbolsToCapture: `
(expression_statement
	(assignment
		left: (identifier) @assignment.name
		right: (identifier) @assignment.content) @assignment)
`,
}

type pyAnalyzer struct {
	*ContextData
}

func (an *pyAnalyzer) FallbackSymbol() Symbol {
	return Builtin("any")
}

func (an *pyAnalyzer) AnalyzeNode(n SyntaxNode) Symbol {
	// TODO:
	return Builtin("void")
}

func (an *pyAnalyzer) AnalyzeImport(params ImportParams) ResolvedImport {
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
}
