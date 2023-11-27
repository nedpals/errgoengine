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

func (pos Position) addUnsafe(pos2 Position) Position {
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

// applyDeleteOperation applies a delete operation to the document based on the changeset
func applyDeleteOperation(doc *EditableDocument, changeset Changeset) Position {
	diff := changeset.EndPos.Index - changeset.StartPos.Index

	// remove the line if the changeset is an empty and the position covers the entire line
	if changeset.StartPos.Column == 0 && changeset.EndPos.Column == len(doc.modifiedLines[changeset.EndPos.Line]) {
		doc.modifiedLines = append(doc.modifiedLines[:changeset.StartPos.Line], doc.modifiedLines[changeset.EndPos.Line+1:]...)
		return Position{
			Line: -1,
			// Column: -diff,
		}
	}

	left := ""
	right := ""

	if changeset.StartPos.Line < len(doc.modifiedLines) {
		targetLeftColumn := min(changeset.StartPos.Column, len(doc.modifiedLines[changeset.StartPos.Line]))
		left = doc.modifiedLines[changeset.StartPos.Line][:targetLeftColumn]

		// get the right part of the line if the changeset start line is the same as the end line
		if changeset.StartPos.Line == changeset.EndPos.Line {
			targetRightColumn := min(changeset.EndPos.Column, len(doc.modifiedLines[changeset.EndPos.Line]))
			right = doc.modifiedLines[changeset.EndPos.Line][targetRightColumn:]
		}
	}

	doc.modifiedLines[changeset.StartPos.Line] = left + right
	return Position{
		Line:   0,
		Column: -diff,
		Index:  -diff,
	}
}

// applyInsertOperation applies an insert operation to the document based on the changeset
func applyInsertOperation(doc *EditableDocument, changeset Changeset) Position {
	diffPosition := Position{}
	left := ""
	right := ""

	shouldAddNewLine := strings.HasSuffix(changeset.NewText, "\n")
	if shouldAddNewLine {
		// remove the newline from the changeset
		changeset.NewText = strings.TrimSuffix(changeset.NewText, "\n")
	}

	// if the changeset start line is not the same as the end line, split it
	if changeset.StartPos.Line < len(doc.modifiedLines) {
		targetLeftColumn := min(changeset.StartPos.Column, len(doc.modifiedLines[changeset.StartPos.Line]))
		left = doc.modifiedLines[changeset.StartPos.Line][:targetLeftColumn]

		// endpos is ignored since we are only using a single position for determining the right side
		targetRightColumn := min(changeset.StartPos.Column, len(doc.modifiedLines[changeset.StartPos.Line]))
		right = doc.modifiedLines[changeset.EndPos.Line][targetRightColumn:]
	}

	// insert the new text
	doc.modifiedLines[changeset.StartPos.Line] = left + changeset.NewText

	// if the changeset has a newline, split the line
	if shouldAddNewLine {
		doc.modifiedLines = append(
			append(
				append([]string{}, doc.modifiedLines[:min(changeset.StartPos.Line+1, len(doc.modifiedLines))]...), // add the lines before the changeset start line
				"", // add an empty line
			),
			doc.modifiedLines[min(changeset.EndPos.Line+1, len(doc.modifiedLines)):]..., // add the lines after the changeset end line
		)

		diffPosition.Line++
		changeset.NewText = strings.TrimSuffix(changeset.NewText, "\n")

		// deduct the length of the left part of the line from the diff position
		diffPosition.Column = -len(left)
		diffPosition.Index = -len(left)
	} else {
		diffPosition.Column = len(changeset.NewText)
		diffPosition.Index = len(changeset.NewText)
	}

	// insert the right part of the line. in the case that a new line was
	// inserted, the right part of the line will be inserted in the next line
	doc.modifiedLines[changeset.StartPos.Line+diffPosition.Line] += right

	return diffPosition
}

func applyOperation(op string, doc *EditableDocument, changeset Changeset) Position {
	// to avoid out of bounds error. limit the endpos column to the length of the doc line
	changeset.EndPos.Column = min(changeset.EndPos.Column, len(doc.modifiedLines[changeset.EndPos.Line]))

	switch op {
	case "insert":
		return applyInsertOperation(doc, changeset)
	case "delete":
		return applyDeleteOperation(doc, changeset)
	case "replace":
		deleteDiff := applyDeleteOperation(doc, Changeset{
			NewText:  "",
			Id:       changeset.Id,
			StartPos: changeset.StartPos,
			EndPos:   changeset.EndPos,
		})

		insertDiff := applyInsertOperation(doc, Changeset{
			NewText:  changeset.NewText,
			Id:       changeset.Id,
			StartPos: changeset.StartPos.Add(deleteDiff),
			EndPos:   changeset.StartPos.Add(deleteDiff),
		})

		// combine the diff to create an interesecting diff
		return insertDiff.addUnsafe(deleteDiff)
	default:
		return Position{}
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

			if line == changeset.StartPos.Line {
				startPos.Column = changeset.StartPos.Column
			}

			if line == changeset.EndPos.Line {
				endPos.Column = changeset.EndPos.Column
			} else if line < len(doc.modifiedLines) {
				// if the line is the last line, set the end position to the length of the line.
				// take note, this is the length of the original line, not the modified line
				endPos.Column = len(doc.modifiedLines[line])
			}

			finalChangeset := Changeset{
				Id:        changeset.Id,
				StartPos:  startPos,
				EndPos:    endPos,
				IsChanged: true, // turn it on so that FillIndex will be called in the earlier part of this function
			}.Add(diffPosition)

			// avoid removals with 0 difference in range
			if finalChangeset.StartPos.Eq(finalChangeset.EndPos) {
				continue
			}

			diffPosition = diffPosition.addUnsafe(
				doc.Apply(finalChangeset),
			)
		}

		return diffPosition
	}

	// if the changeset is a newline, split the new text into lines and apply them one by one
	nlCount := strings.Count(changeset.NewText, "\n")
	hasTrailingNewLine := strings.HasSuffix(changeset.NewText, "\n")
	if (nlCount == 1 && !hasTrailingNewLine) || nlCount > 1 {
		newLines := strings.Split(changeset.NewText, "\n")

		for i, line := range newLines {
			textToAdd := line
			// the extra empty string or a last string will
			// indicate that the will be inserted without newline
			if i < len(newLines)-1 {
				textToAdd += "\n"
			}

			startPos := Position{Line: changeset.StartPos.Line, Column: 0}
			endPos := Position{Line: changeset.EndPos.Line, Column: 0}

			if i == 0 {
				startPos.Column = changeset.StartPos.Column
			}

			if endPos.Line == changeset.EndPos.Line {
				endPos.Column = changeset.EndPos.Column
			} else if i < len(newLines)-1 && changeset.StartPos.Line+i < len(doc.modifiedLines) {
				endPos.Column = len(doc.modifiedLines[changeset.StartPos.Line+i])
			}

			diffPosition = diffPosition.addUnsafe(
				doc.Apply(
					Changeset{
						Id:       changeset.Id,
						NewText:  textToAdd,
						StartPos: startPos,
						EndPos:   endPos,
					}.Add(diffPosition),
				),
			)
		}

		return diffPosition
	}

	selectedOperation := "insert"

	// if the changeset has a definite position range, check if it is a replacement or a deletion
	if !changeset.StartPos.Eq(changeset.EndPos) {
		if len(changeset.NewText) == 0 {
			// if the changeset text is empty but has a definite position
			// range, this means that the changeset is a deletion
			selectedOperation = "delete"
		} else if !hasTrailingNewLine {
			// if the changeset has no trailing newline, this
			// means that the changeset is a replacement
			selectedOperation = "replace"
		}
	}

	// apply editing operation
	diffPosition = diffPosition.addUnsafe(
		applyOperation(
			selectedOperation,
			doc,
			changeset.Add(diffPosition),
		),
	)

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
