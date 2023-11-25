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

func (pos Position) IsInBetween(loc Location) bool {
	return pos.Index >= loc.StartPos.Index && pos.Index <= loc.EndPos.Index
}

func (pos Position) Add(pos2 Position) Position {
	return Position{
		Line:   max(pos.Line+pos2.Line, 0),
		Column: max(pos.Column+pos2.Column, 0),
		Index:  max(pos.Index+pos2.Index, 0),
	}
}

func (pos Position) addNoCheck(pos2 Position) Position {
	return Position{
		Line:   pos.Line + pos2.Line,
		Column: pos.Column + pos2.Column,
		Index:  pos.Index + pos2.Index,
	}
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

func (loc Location) IsWithin(other Location) bool {
	return loc.StartPos.IsInBetween(other) && loc.EndPos.IsInBetween(other)
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
	Id        int
	NewText   string
	StartPos  Position
	EndPos    Position
	IsChanged bool
}

func (c Changeset) Add(posToAdd Position) Changeset {
	return Changeset{
		Id:        c.Id,
		NewText:   c.NewText,
		StartPos:  c.StartPos.Add(posToAdd),
		EndPos:    c.EndPos.Add(posToAdd),
		IsChanged: true,
	}
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

func (doc *EditableDocument) Apply(changeset Changeset) Position {
	diffPosition := Position{}

	// add id if not present
	if changeset.Id == 0 {
		doc.currentId++
		changeset.Id = doc.currentId
	}

	if changeset.IsChanged {
		changeset.StartPos = doc.FillIndex(changeset.StartPos)
		changeset.EndPos = doc.FillIndex(changeset.EndPos)
		changeset.IsChanged = false
	}

	// if the changeset text is empty but has a definite position range, this means that the changeset is a deletion
	if len(changeset.NewText) == 0 && changeset.StartPos.Line != changeset.EndPos.Line {
		// if the changeset start line is not the same as the end line, split it
		for line := changeset.StartPos.Line; line <= changeset.EndPos.Line; line++ {
			startPos := Position{Line: line, Column: 0}
			endPos := Position{Line: line, Column: 0}

			if line < len(doc.modifiedLines) {
				// if the line is the last line, set the end position to the length of the line.
				// take note, this is the length of the original line, not the modified line
				endPos.Column = len(doc.modifiedLines[line])
			}

			if line == changeset.StartPos.Line {
				startPos.Column = changeset.StartPos.Column
			} else if line == changeset.EndPos.Line {
				endPos.Column = changeset.EndPos.Column
			}

			diffPosition = diffPosition.addNoCheck(doc.Apply(Changeset{
				Id:        changeset.Id,
				StartPos:  startPos,
				EndPos:    endPos,
				IsChanged: true, // turn it on so that FillIndex will be called in the earlier part of this function
			}.Add(diffPosition)))
		}

		return diffPosition
	}

	// if the changeset is a newline, split the new text into lines and apply them
	nlCount := strings.Count(changeset.NewText, "\n")
	if (nlCount == 1 && !strings.HasSuffix(changeset.NewText, "\n")) || nlCount > 1 {
		newLines := strings.Split(changeset.NewText, "\n")

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

			diffPosition = diffPosition.addNoCheck(doc.Apply(Changeset{
				Id:       changeset.Id,
				NewText:  textToAdd,
				StartPos: startPos,
				EndPos:   endPos,
			}.Add(diffPosition)))
		}

		return diffPosition
	}

	// to avoid out of bounds error. limit the endpos column to the length of the doc line
	changeset.EndPos.Column = min(max(changeset.EndPos.Column, 0), len(doc.modifiedLines[changeset.EndPos.Line]))

	if len(changeset.NewText) == 0 && changeset.StartPos.Column == 0 && changeset.EndPos.Column == len(doc.modifiedLines[changeset.EndPos.Line]) {
		// remove the line if the changeset is an empty and the position covers the entire line
		doc.modifiedLines = append(doc.modifiedLines[:changeset.StartPos.Line], doc.modifiedLines[changeset.EndPos.Line+1:]...)
		diffPosition.Line = -1
		diffPosition.Index = -(changeset.EndPos.Index - changeset.StartPos.Index)
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
			diffPosition.Line = 1
			diffPosition.Index = len(changeset.NewText)
		} else {
			doc.modifiedLines[changeset.StartPos.Line] = left + changeset.NewText + right
			diffPosition.Line = 0
			diffPosition.Column = len(changeset.NewText) - (changeset.EndPos.Column - changeset.StartPos.Column)
			diffPosition.Index = len(changeset.NewText) - (changeset.EndPos.Index - changeset.StartPos.Index)
		}
	}

	// fmt.Printf("new: %v for %q\n", changeset, doc.modifiedLines)

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
			Row:    uint32(changeset.EndPos.Line - diffPosition.Line),
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
		return Position{}
	}

	doc.tree = newTree
	return diffPosition
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
		return list[:to+1]
	}

	return list[from : to+1]
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
