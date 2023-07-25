package main

import (
	"strconv"
)

var NullPointerException = ErrorTemplate{
	Name:              "NullPointerException",
	Pattern:           `Exception in thread "(?P<thread>\w+)" java\.lang\.NullPointerException`,
	StackTracePattern: `\s+at (?P<symbol>\S+)\((?P<path>\S+):(?P<position>\d+)\)`,
	LocationConverterFn: func(path, pos string) Location {
		trueLine, err := strconv.Atoi(pos)
		if err != nil {
			panic(err)
		}

		return Location{
			DocumentPath: path,
			Position:     Position{Line: trueLine},
		}
	},
	OnGenExplainFn: func(doc *Document, cd *ContextData) string {
		return "Your program try to access or manipulate an object reference that is currently pointing to `null`, meaning it doesn't refer to any actual object in memory. This typically happens when you forget to initialize an object before using it, or when you try to access an object that hasn't been properly assigned a value. "
	},
	OnGenBugFixFn: func(doc *Document, cd *ContextData) []BugFix {
		return []BugFix{}
	},
}
