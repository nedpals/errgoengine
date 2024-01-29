package java

import (
	"fmt"
	"strings"
	"unicode"

	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/java"
	sitter "github.com/smacker/go-tree-sitter"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	// Runtime
	errorTemplates.MustAdd(java.Language, NullPointerException)
	errorTemplates.MustAdd(java.Language, ArrayIndexOutOfBoundsException)
	errorTemplates.MustAdd(java.Language, ArithmeticException)
	errorTemplates.MustAdd(java.Language, NegativeArraySizeException)
	errorTemplates.MustAdd(java.Language, StringIndexOutOfBoundsException)
	errorTemplates.MustAdd(java.Language, NoSuchElementException)
	errorTemplates.MustAdd(java.Language, NumberFormatException)
	errorTemplates.MustAdd(java.Language, InputMismatchException)

	// Compile time
	errorTemplates.MustAdd(java.Language, PublicClassFilenameMismatchError)
	errorTemplates.MustAdd(java.Language, ParseEndOfFileError)
	errorTemplates.MustAdd(java.Language, UnreachableStatementError)
	errorTemplates.MustAdd(java.Language, ArrayRequiredTypeError)
	errorTemplates.MustAdd(java.Language, SymbolNotFoundError)
	errorTemplates.MustAdd(java.Language, NonStaticMethodAccessError)
	errorTemplates.MustAdd(java.Language, UnclosedCharacterLiteralError)
	errorTemplates.MustAdd(java.Language, OperatorCannotBeAppliedError)
	errorTemplates.MustAdd(java.Language, PrecisionLossError)
	errorTemplates.MustAdd(java.Language, MissingReturnError)
	errorTemplates.MustAdd(java.Language, NotAStatementError)
	errorTemplates.MustAdd(java.Language, IncompatibleTypesError)
	errorTemplates.MustAdd(java.Language, UninitializedVariableError)
	errorTemplates.MustAdd(java.Language, AlreadyDefinedError)
	errorTemplates.MustAdd(java.Language, PrivateAccessError)
	errorTemplates.MustAdd(java.Language, IllegalExpressionStartError)
	errorTemplates.MustAdd(java.Language, UnclosedStringLiteralError)
	errorTemplates.MustAdd(java.Language, CannotBeAppliedError)
	errorTemplates.MustAdd(java.Language, InvalidMethodDeclarationError)
	errorTemplates.MustAdd(java.Language, IdentifierExpectedError)
	errorTemplates.MustAdd(java.Language, IllegalCharacterError)
	errorTemplates.MustAdd(java.Language, CharacterExpectedError)
}

func runtimeErrorPattern(errorName string, pattern string) string {
	p := fmt.Sprintf(
		`Exception in thread "(?P<thread>\w+)" %s`,
		strings.ReplaceAll(errorName, ".", `\.`),
	)
	if len(pattern) != 0 {
		p += ": " + pattern
	}
	return p
}

const comptimeStackTracePattern = `(?P<path>\S+):(?P<position>\d+)`

func comptimeErrorPattern(pattern string, endPattern_ ...string) string {
	endPattern := ".*"
	if len(endPattern_) != 0 {
		endPattern = `(?:.|\s)+` + endPattern_[0]
	}
	return fmt.Sprintf(`$stacktrace: error: %s%s`, pattern, endPattern)
}

// TODO:
func getIdentifierNode(node lib.SyntaxNode) lib.SyntaxNode {
	currentNode := node
	for currentNode.Type() != "identifier" {
		return currentNode
	}
	return currentNode
}

func getDefaultValueForType(sym lib.Symbol) string {
	switch sym {
	case java.BuiltinTypes.Integral.IntSymbol:
		return "0"
	case java.BuiltinTypes.Integral.LongSymbol:
		return "0L"
	case java.BuiltinTypes.Integral.ShortSymbol:
		return "0"
	case java.BuiltinTypes.FloatingPoint.DoubleSymbol:
		return "0.0"
	case java.BuiltinTypes.StringSymbol:
		return "\"example\""
	default:
		return "null"
	}
}

func symbolToValueNodeType(sym lib.Symbol) []string {
	switch sym {
	case java.BuiltinTypes.Integral.IntSymbol:
		return []string{"decimal_integer_literal"}
	case java.BuiltinTypes.FloatingPoint.DoubleSymbol:
		return []string{"decimal_floating_point_literal"}
	case java.BuiltinTypes.BooleanSymbol:
		return []string{"true", "false"}
	case java.BuiltinTypes.Integral.CharSymbol:
		return []string{"character_literal"}
	case java.BuiltinTypes.StringSymbol:
		return []string{"string_literal"}
	case java.BuiltinTypes.VoidSymbol, java.BuiltinTypes.NullSymbol:
		return []string{"null_literal"}
	default:
		return []string{}
	}
}

func castValueNode(node lib.SyntaxNode, targetSym lib.Symbol) string {
	switch targetSym {
	case java.BuiltinTypes.Integral.IntSymbol:
		switch node.Type() {
		case "character_literal", "decimal_floating_point_literal":
			return fmt.Sprintf("(int) '%s'", node.Text())
		case "string_literal":
			return fmt.Sprintf("Integer.parseInt(%s)", node.Text())
		default:
			return node.Text()
		}
	default:
		return node.Text()
	}
}

func wrapStatement(s *lib.BugFixStep, header string, footer string, loc lib.Location, withTrailingNewLine bool) lib.Position {
	line := s.Doc.ModifiedLineAt(loc.StartPos.Line)
	startCol, endCol := GetSpaceBoundary(line, loc.StartPos.Column, loc.StartPos.Column)
	spaces := line[startCol:endCol]
	// fmt.Printf("%q %d %d %q\n", line, startCol, endCol, spaces)

	s.AddFix(lib.FixSuggestion{
		NewText:       spaces + header + "\n" + getIndent(spaces, 1),
		StartPosition: loc.StartPos,
		EndPosition:   loc.StartPos,
	})

	// get spaces from the first index to the non-space boundary
	// in order to get the correct indentation
	if len(spaces) == 0 {
		// get spaces from the start line
		spaces = getSpaceFromBeginning2(line, loc.StartPos.Line, len(line))
	}

	footer = "\n" + footer
	if withTrailingNewLine {
		footer += "\n"
	}

	indent := getIndent(spaces, 1)

	s.AddFix(lib.FixSuggestion{
		NewText:       strings.Replace(strings.Replace(footer, "<i>", indent, -1), "\t", spaces, -1),
		StartPosition: loc.EndPos,
		EndPosition:   loc.EndPos,
	})

	return s.DiffPosition
}

type spaceComputeDirection int

const (
	spaceComputeDirectionLeft spaceComputeDirection = iota
	spaceComputeDirectionRight
)

func isSpace(line string, idx int) bool {
	if idx >= len(line) || idx < 0 {
		return false
	}
	return unicode.IsSpace(rune(line[idx]))
}

func getSpaceBoundaryIndiv(line string, idx int, defaultDirection spaceComputeDirection) int {
	stopOnSpace := true
	if isSpace(line, idx) {
		stopOnSpace = false
	}

	if defaultDirection == spaceComputeDirectionLeft {
		if idx-1 >= 0 && ((!stopOnSpace && !isSpace(line, idx-1)) ||
			(stopOnSpace && isSpace(line, idx-1))) {
			return idx
		}

		if idx-1 < 0 {
			// check if the current index is not a space
			if !isSpace(line, idx) {
				// go to the reverse direction to get the nearest space
				newIdx := getSpaceBoundaryIndiv(line, idx, spaceComputeDirectionRight)
				return newIdx
			}

			return idx
		}

		for idx > 0 {
			if (stopOnSpace && isSpace(line, idx)) ||
				(!stopOnSpace && !isSpace(line, idx)) {
				break
			}
			idx--
		}
	}

	if defaultDirection == spaceComputeDirectionRight {
		if idx == len(line) {
			// go to the last character of the line
			idx--
		}

		if idx+1 < len(line)-1 && ((!stopOnSpace && !isSpace(line, idx+1)) ||
			(stopOnSpace && isSpace(line, idx+1))) {
			return idx
		}

		// check if the current index is not a space
		if !isSpace(line, idx) {
			// go to the reverse direction to get the nearest space
			newIdx := getSpaceBoundaryIndiv(line, idx, spaceComputeDirectionLeft)
			if newIdx+1 < len(line) {
				return newIdx + 1
			}
			return newIdx
		}

		for idx < len(line) {
			if (stopOnSpace && isSpace(line, idx)) ||
				(!stopOnSpace && !isSpace(line, idx)) {
				break
			}
			idx++
		}
	}

	return idx
}

func GetSpaceBoundary(line string, from int, to int) (int, int) {
	if len(line) == 0 {
		return 0, 0
	}

	if from > to {
		from, to = to, from
	}

	from = getSpaceBoundaryIndiv(line, from, spaceComputeDirectionLeft)
	if from > 0 && from == to && isSpace(line, from-1) {
		from = getSpaceBoundaryIndiv(line, from-1, spaceComputeDirectionLeft)
	}

	to = getSpaceBoundaryIndiv(line, to, spaceComputeDirectionRight)

	if from != to {
		// check if there are still non-space characters in the range
		for i := from; i < to; i++ {
			if isSpace(line, i) {
				continue
			}

			// return the nearest space boundary
			return from, i
		}
	}

	return from, to
}

func getSpace(lineStr string, line int, from int, to int) string {
	startCol, endCol := GetSpaceBoundary(lineStr, from, to)
	if startCol > endCol {
		startCol, endCol = endCol, startCol
	}

	return lineStr[startCol:endCol]
}

func getSpaceFromBeginning(doc *lib.Document, line int, to int) string {
	return getSpaceFromBeginning2(doc.LineAt(line), line, to)
}

func getSpaceFromBeginning2(lines string, line int, to int) string {
	return getSpace(lines, line, 0, to)
}

func getIndent(spaces string, by int) string {
	if len(spaces) < 4 {
		return spaces
	}

	indent := spaces
	if len(indent) > 4 {
		indent = spaces[:4]
	}

	if by == 1 {
		return indent
	}

	return strings.Repeat(indent, by)
}

func indentSpace(spaces string, by int) string {
	indent := getIndent(spaces, by)
	return spaces + indent
}

func nearestMissingNodeFromPos(cursor *sitter.TreeCursor, pos lib.Position) *sitter.Node {
	defer cursor.GoToParent()

	// hope it executes to avoid stack overflow
	if !cursor.GoToFirstChild() {
		return nil
	}

	for {
		currentNode := cursor.CurrentNode()
		pointA := currentNode.StartPoint()
		pointB := currentNode.EndPoint()

		if uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
			if currentNode.IsMissing() {
				return currentNode
			} else if currentNode.ChildCount() != 0 {
				if gotNode := nearestMissingNodeFromPos(cursor, pos); gotNode != nil {
					return gotNode
				}
			}
		}

		if !cursor.GoToNextSibling() {
			return nil
		}
	}
}
