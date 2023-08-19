package java

import (
	"github.com/nedpals/errgoengine/languages/java"
	"github.com/nedpals/errgoengine/lib"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	errorTemplates.Add(java.Language, NullPointerException)
	errorTemplates.Add(java.Language, ArrayIndexOutOfBoundsException)
	errorTemplates.Add(java.Language, ArithmeticException)
	errorTemplates.Add(java.Language, PublicClassFilenameMismatchError)
}
