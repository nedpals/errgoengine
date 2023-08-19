package java

import (
	"github.com/nedpals/errgoengine/languages/java"
	"github.com/nedpals/errgoengine/lib"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	errorTemplates.Add(java.Language, NullPointerException)
}
