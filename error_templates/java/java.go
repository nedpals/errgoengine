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

func wrapWithCondStatement(s *lib.BugFixStep, d *lib.Document, condType string, cond string, loc lib.Location, withNewLine bool) {
	if len(condType) == 0 {
		condType = "if"
	}

	line := d.LineAt(loc.StartPos.Line)
	startCol, endCol := getSpaceBoundary(line, loc.StartPos.Column, loc.StartPos.Column, true)
	spaces := line[startCol:endCol]

	opening := fmt.Sprintf("%s (%s)", condType, cond)
	if condType == "else" || len(cond) == 0 {
		if startCol == loc.StartPos.Column && condType == "else" {
			opening = " else"
		} else {
			opening = condType
		}
	}

	s.AddFix(lib.FixSuggestion{
		NewText: spaces + opening + " {\n" + getIndent(spaces, 1),
		StartPosition: lib.Position{
			Line:   loc.StartPos.Line,
			Column: startCol,
		},
		EndPosition: lib.Position{
			Line:   loc.StartPos.Line,
			Column: startCol,
		},
	})

	// get spaces from the first index to the non-space boundary
	// in order to get the correct indentation
	if len(spaces) == 0 {
		// get spaces from the start line
		startCol, endCol := getSpaceBoundary(line, 0, len(line), false)
		spaces = line[startCol:endCol]
	}

	closing := "\n" + spaces + "}"
	if withNewLine {
		closing += "\n"
	}

	s.AddFix(lib.FixSuggestion{
		NewText:       closing,
		StartPosition: loc.EndPos,
		EndPosition:   loc.EndPos,
	})
}

func getSpaceBoundary(line string, from int, to int, reverse bool) (int, int) {
	if from > to {
		from, to = to, from
	}

	for i := to - 1; i >= 0; i-- {
		if !unicode.IsSpace(rune(line[i])) {
			break
		}
		to = i
	}

	for i := from; i < to; i++ {
		if !unicode.IsSpace(rune(line[i])) {
			break
		}
		from++
	}

	return from, to
}

func getSpace(doc *lib.Document, line int, from int, to int, reverse bool) string {
	startCol, endCol := getSpaceBoundary(doc.LineAt(from), from, to, reverse)
	if startCol > endCol {
		startCol, endCol = endCol, startCol
	}

	lineStr := doc.LineAt(line)
	fmt.Println(startCol, endCol, lineStr)
	return lineStr[startCol:endCol]
}

func getSpaceFromBeginning(doc *lib.Document, line int, to int) string {
	return getSpace(doc, line, 0, to, true)
}

func getIndent(spaces string, by int) string {
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
