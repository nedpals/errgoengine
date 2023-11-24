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
	Id         int
	NewText    string
	StartPos   Position
	EndPos     Position
	linesAdded int
	IsChanged  bool
}

func (c Changeset) AddLines(lines int) Changeset {
	if lines != 0 {
		fmt.Println("adding lines", lines)

		c.linesAdded += lines
		c.StartPos.Line += lines
		c.EndPos.Line += lines
		c.IsChanged = true
	}
	return c
}

type EditableDocument struct {
	*Document
	tree          *sitter.Tree
	currentId     int
	modifiedLines []string
	parser        *sitter.Parser
	changesets    []Changeset
}

func NewEditableDocument(doc *Document) *EditableDocument {
	parser := sitter.NewParser()
	parser.SetLanguage(doc.Language.SitterLanguage)

	editableDoc := &EditableDocument{
		Document: doc,
		parser:   parser,
	}

	editableDoc.Reset()
	return editableDoc
}

func (doc *EditableDocument) FillIndex(pos Position) Position {
	if pos.Line == 0 && pos.Column == 0 && pos.Index == 0 {
		return pos
	}

	gotIdx := 0
	for lIdx, line := range doc.modifiedLines {
		if lIdx == pos.Line {
			break
		}

		gotIdx += len(line) + 1
	}

	if pos.Line >= len(doc.modifiedLines) {
		gotIdx += (pos.Line - len(doc.modifiedLines))
	}

	gotIdx += pos.Column
	return Position{
		Line:   pos.Line,
		Column: pos.Column,
		Index:  gotIdx,
	}
}

func (doc *EditableDocument) Apply(changeset Changeset) int {
	totalLinesAdded := 0

	// add id if not present
	if changeset.Id == 0 {
		doc.currentId++
		changeset.Id = doc.currentId
	}

	if changeset.IsChanged {
		changeset.StartPos.Index = doc.FillIndex(changeset.StartPos).Index
		changeset.EndPos.Index = doc.FillIndex(changeset.EndPos).Index
		changeset.IsChanged = false
	}

	// if the changeset text is empty but has a definite position range, this means that the changeset is a deletion
	if len(changeset.NewText) == 0 {
		total := 0

		// if the changeset start line is not the same as the end line, split it
		for i := changeset.StartPos.Line; i < len(doc.modifiedLines); {
			if i > changeset.EndPos.Line {
				break
			}

			startPos := Position{Line: i, Column: 0}
			endPos := Position{Line: i, Column: len(doc.modifiedLines[i])}

			if i == changeset.StartPos.Line {
				startPos.Column = changeset.StartPos.Column
			} else if i == changeset.EndPos.Line {
				endPos.Column = changeset.EndPos.Column
			}

			if i != changeset.StartPos.Line {
				startPos.Line = i - 1
				endPos.Line = i - 1
			}

			changeset := Changeset{
				Id:       changeset.Id,
				NewText:  changeset.NewText,
				StartPos: doc.FillIndex(startPos),
				EndPos:   doc.FillIndex(endPos),
			}

			total += doc.Apply(changeset.AddLines(total))
		}

		return total
	}

	// if the changeset is a newline, split the new text into lines and apply them
	nlCount := strings.Count(changeset.NewText, "\n")
	if (nlCount == 1 && !strings.HasSuffix(changeset.NewText, "\n")) || nlCount > 1 {
		newLines := strings.Split(changeset.NewText, "\n")
		total := 0

		for i, line := range newLines {
			textToAdd := line
			if i < len(newLines)-1 {
				textToAdd += "\n"
			}

			startPos := Position{Line: changeset.StartPos.Line, Column: changeset.StartPos.Column}
			endPos := Position{Line: changeset.EndPos.Line, Column: len(doc.modifiedLines[changeset.StartPos.Line+i])}
			if i > 0 {
				startPos.Column = 0
			}

			if i == len(newLines)-1 {
				endPos.Column = changeset.EndPos.Column
			}

			changeset := Changeset{
				Id:       changeset.Id,
				NewText:  textToAdd,
				StartPos: startPos,
				EndPos:   endPos,
			}

			total += doc.Apply(changeset.AddLines(total))
		}

		return total
	}

	// to avoid out of bounds error. limit the endpos column to the length of the doc line
	changeset.EndPos.Column = min(changeset.EndPos.Column, len(doc.modifiedLines[changeset.EndPos.Line]))

	if len(changeset.NewText) == 0 && changeset.StartPos.Column == 0 && changeset.EndPos.Column == len(doc.modifiedLines[changeset.StartPos.Line]) {
		// remove the line if the changeset is an empty and the position covers the entire line
		doc.modifiedLines = append(doc.modifiedLines[:changeset.StartPos.Line], doc.modifiedLines[changeset.EndPos.Line:]...)
		changeset.linesAdded = -1
	} else {
		left := doc.modifiedLines[changeset.StartPos.Line][:changeset.StartPos.Column]
		right := doc.modifiedLines[changeset.EndPos.Line][changeset.EndPos.Column:]

		// remove newline if the changeset is a newline
		if nlCount >= 1 {
			changeset.NewText = changeset.NewText[:len(changeset.NewText)-1]
		}

		// create a new line if the changeset is a newline
		if len(changeset.NewText) > 1 && nlCount >= 1 {
			doc.modifiedLines = append(
				append(append([]string{}, doc.modifiedLines[:changeset.StartPos.Line]...), ""),
				doc.modifiedLines[changeset.EndPos.Line:]...)

			doc.modifiedLines[changeset.StartPos.Line] = left + changeset.NewText
			changeset.linesAdded++
		} else {
			fmt.Println(changeset.EndPos.Line)

			doc.modifiedLines[changeset.StartPos.Line] = left + changeset.NewText + right
			changeset.linesAdded = 0
		}
	}

	fmt.Printf("new: %q for %q\n", changeset.NewText, doc.modifiedLines)

	// add changeset
	doc.changesets = append(doc.changesets, changeset)

	// reparse the document
	doc.tree.Edit(sitter.EditInput{
		StartIndex:  uint32(changeset.StartPos.Index),
		OldEndIndex: uint32(changeset.EndPos.Index),
		NewEndIndex: uint32(changeset.StartPos.Index + len(changeset.NewText)),
		StartPoint: sitter.Point{
			Row:    uint32(changeset.StartPos.Line),
			Column: uint32(changeset.StartPos.Column),
		},
		OldEndPoint: sitter.Point{
			Row:    uint32(changeset.EndPos.Line - totalLinesAdded),
			Column: uint32(changeset.EndPos.Column),
		},
		NewEndPoint: sitter.Point{
			Row:    uint32(changeset.StartPos.Line),
			Column: uint32(changeset.StartPos.Column + len(changeset.NewText)),
		},
	})

	newTree, err := doc.parser.ParseCtx(
		context.Background(),
		doc.tree,
		[]byte(doc.String()),
	)
	if err != nil {
		return 0
	}

	doc.tree = newTree
	return totalLinesAdded
}

func (doc *EditableDocument) String() string {
	return strings.Join(doc.modifiedLines, "\n")
}

func (doc *EditableDocument) ModifiedLineAt(idx int) string {
	if idx < 0 || idx >= len(doc.modifiedLines) {
		return ""
	}
	return doc.modifiedLines[idx]
}

func linesAt(list []string, from int, to int) []string {
	if from > to {
		from, to = to, from
	}
	if to == -1 {
		to = len(list)
	}
	from = max(from, 0)
	to = min(to, len(list))
	if from == 0 && to == len(list) {
		return list
	} else if from > 0 && to == len(list) {
		return list[from:]
	} else if from == 0 && to < len(list) {
		return list[:to]
	}

	return list[from:to]
}

func (doc *EditableDocument) ModifiedLinesAt(from int, to int) []string {
	return linesAt(doc.modifiedLines, from, to)
}

func (doc *EditableDocument) Reset() {
	rawLines := doc.Lines()
	lines := make([]string, len(rawLines))
	copy(lines, rawLines)

	doc.modifiedLines = lines
	doc.changesets = nil
	doc.tree = doc.Tree.Copy()
	doc.parser.Reset()
}

type Document struct {
	Path        string
	Contents    string
	cachedLines []string
	Language    *Language
	Tree        *sitter.Tree
}

func (doc *Document) Editable() *EditableDocument {
	return NewEditableDocument(doc)
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
	if doc.cachedLines == nil || (len(doc.Contents) != 0 && len(doc.cachedLines) == 0) {
		doc.cachedLines = strings.Split(doc.Contents, "\n")
	}
	return linesAt(doc.cachedLines, from, to)
}

func (doc *Document) Lines() []string {
	return doc.LinesAt(-2, -1)
}

func (doc *Document) TotalLines() int {
	return len(doc.Lines())
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
