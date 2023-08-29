package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/java"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	// Runtime
	errorTemplates.Add(java.Language, NullPointerException)
	errorTemplates.Add(java.Language, ArrayIndexOutOfBoundsException)
	errorTemplates.Add(java.Language, ArithmeticException)

	// Compile time
	errorTemplates.Add(java.Language, PublicClassFilenameMismatchError)
	errorTemplates.Add(java.Language, ParseEndOfFileError)
	errorTemplates.Add(java.Language, UnreachableStatementError)
	errorTemplates.Add(java.Language, ArrayRequiredTypeError)
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
