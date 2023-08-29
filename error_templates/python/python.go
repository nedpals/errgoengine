package python

import (
	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/python"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	// Runtime error
	errorTemplates.Add(python.Language, ZeroDivisionError)
	errorTemplates.Add(python.Language, NameError)
	errorTemplates.Add(python.Language, ValueError)
	errorTemplates.Add(python.Language, AttributeError)

	// Compile time error
	errorTemplates.Add(python.Language, SyntaxError)
	errorTemplates.Add(python.Language, IndentationError)
}

func compileTimeError(pattern string) string {
	return lib.CustomErrorPattern("$stacktrace" + pattern)
}
