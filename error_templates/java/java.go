package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/java"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	// Runtime
	errorTemplates.MustAdd(java.Language, NullPointerException)
	errorTemplates.MustAdd(java.Language, ArrayIndexOutOfBoundsException)
	errorTemplates.MustAdd(java.Language, ArithmeticException)
	errorTemplates.MustAdd(java.Language, NegativeArraySizeException)

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
	errorTemplates.MustAdd(java.Language, BracketMismatchError)
	errorTemplates.MustAdd(java.Language, InvalidMethodDeclarationError)
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
