package errgoengine

import (
	"context"
	"fmt"
	"io"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type Position struct {
	Line   int
	Column int
	Index  int
}

func (a Position) Eq(b Position) bool {
	return a.Column == b.Column &&
		a.Line == b.Line &&
		a.Index == b.Index
}

func (pos Position) String() string {
	return fmt.Sprintf("[%d,%d | %d]", pos.Line, pos.Column, pos.Index)
}

type Location struct {
	DocumentPath string
	// Position
	StartPos Position
	EndPos   Position
}

func (loc Location) Point() sitter.Point {
	return sitter.Point{
		Row:    uint32(loc.StartPos.Line),
		Column: uint32(loc.StartPos.Column),
	}
}

func (loc Location) Range() sitter.Range {
	return sitter.Range{
		StartPoint: sitter.Point{
			Row:    uint32(loc.StartPos.Line),
			Column: uint32(loc.StartPos.Column),
		},
		EndPoint: sitter.Point{
			Row:    uint32(loc.EndPos.Line),
			Column: uint32(loc.EndPos.Column),
		},
	}
}

type Changeset struct {
	NewText   string
	StartPos  Position
	EndPos    Position
	IsReplace bool
}

type TempDocument struct {
	*Document
	lines      []string
	changesets []Changeset
}

type Document struct {
	Path        string
	Contents    string
	cachedLines []string
	Language    *Language
	Tree        *sitter.Tree
}

func (doc *TempDocument) ApplyEdit(newText string, startPos Position, endPos Position) {
	isReplace := !startPos.Eq(endPos)
	doc.changesets = append(doc.changesets, Changeset{
		NewText:   newText,
		StartPos:  startPos,
		EndPos:    endPos,
		IsReplace: isReplace,
	})

	if isReplace {
		if startPos.Column == 0 && endPos.Column == 0 {
			doc.cachedLines[startPos.Line] = newText
		} else {
			line := startPos.Line
			left := doc.cachedLines[line][:startPos.Column]
			right := doc.cachedLines[line][endPos.Column:]
			doc.cachedLines[line] = left + newText + right
		}
	} else {
		doc.cachedLines = append(doc.cachedLines[:startPos.Line+1], doc.cachedLines[endPos.Line:]...)
		doc.cachedLines[startPos.Line] = newText
	}
}

func (doc *Document) CreateTempDoc() TempDocument {
	lines := make([]string, len(doc.Lines()))
	copy(lines, doc.Lines())

	return TempDocument{
		Document: doc,
		lines:    lines,
	}
}

func (doc *Document) LineAt(idx int) string {
	if doc.cachedLines == nil {
		doc.cachedLines = strings.Split(doc.Contents, "\n")
	}
	if idx < 0 || idx >= len(doc.cachedLines) {
		return ""
	}
	return doc.cachedLines[idx]
}

func (doc *Document) LinesAt(from int, to int) []string {
	if doc.cachedLines == nil {
		doc.cachedLines = strings.Split(doc.Contents, "\n")
	}
	if from > to {
		from, to = to, from
	}
	from = max(from, 0)
	to = min(to, len(doc.cachedLines))
	if from == 0 && to == len(doc.cachedLines) {
		return doc.cachedLines
	} else if from > 0 && to == len(doc.cachedLines) {
		return doc.cachedLines[from:]
	} else if from == 0 && to < len(doc.cachedLines) {
		return doc.cachedLines[:to]
	}
	return doc.cachedLines[from:to]
}

func (doc *Document) Lines() []string {
	return doc.LinesAt(0, len(doc.cachedLines)-1)
}

func (doc *Document) TotalLines() int {
	return len(doc.cachedLines)
}

func ParseDocument(path string, r io.Reader, parser *sitter.Parser, selectLang *Language, existingDoc *Document) (*Document, error) {
	inputBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	defer parser.Reset()
	parser.SetLanguage(selectLang.SitterLanguage)

	var existingTree *sitter.Tree
	if existingDoc != nil {
		existingTree = existingDoc.Tree
	}

	tree, err := parser.ParseCtx(context.Background(), existingTree, inputBytes)
	if err != nil {
		return nil, err
	}

	if existingDoc != nil {
		existingDoc.Contents = string(inputBytes)
		existingDoc.Tree = tree
		return existingDoc, nil
	}

	return &Document{
		Path:     path,
		Language: selectLang,
		Contents: string(inputBytes),
		Tree:     tree,
	}, nil
}
