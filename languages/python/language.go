package python

import (
	"github.com/nedpals/errgoengine/lib"
	"github.com/smacker/go-tree-sitter/python"
)

var Language = &lib.Language{
	Name:              "Python",
	FilePatterns:      []string{".py"},
	SitterLanguage:    python.GetLanguage(),
	StackTracePattern: `\s+File "(?P<path>\S+)", line (?P<position>\d+), in (?P<symbol>\S+)`,
	ErrorPattern:      `Traceback \(most recent call last\):$stacktrace$message`,
}
