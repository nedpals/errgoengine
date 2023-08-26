package languages

import (
	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/java"
	"github.com/nedpals/errgoengine/languages/python"
)

var SupportedLanguages = []*lib.Language{
	java.Language,
	python.Language,
}
