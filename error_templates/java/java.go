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

	// Compile time
	errorTemplates.MustAdd(java.Language, PublicClassFilenameMismatchError)
	errorTemplates.MustAdd(java.Language, ParseEndOfFileError)
	errorTemplates.MustAdd(java.Language, UnreachableStatementError)
	errorTemplates.MustAdd(java.Language, ArrayRequiredTypeError)
	errorTemplates.MustAdd(java.Language, SymbolNotFoundError)
	errorTemplates.MustAdd(java.Language, NonStaticMethodAccessError)
	errorTemplates.MustAdd(java.Language, UnclosedCharacterLiteralError)
	errorTemplates.MustAdd(java.Language, OperatorCannotBeAppliedError)
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
