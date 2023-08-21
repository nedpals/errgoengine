package lib

import (
	"regexp"
	"strings"
)

type ErrorTemplate struct {
	Name              string
	Pattern           string
	StackTracePattern string
	OnGenExplainFn    GenExplainFn
	OnGenBugFixFn     GenBugFixFn
}

type CompiledErrorTemplate struct {
	ErrorTemplate
	Language          *Language
	Pattern           *regexp.Regexp
	StackTracePattern *regexp.Regexp
}

type ErrorTemplates []*CompiledErrorTemplate

const defaultStackTraceRegex = `(?P<stacktrace>(?:.|\s)*)`

func (tmps *ErrorTemplates) Add(language *Language, template ErrorTemplate) *CompiledErrorTemplate {
	var stackTracePattern *regexp.Regexp

	patternForCompile := ""
	if len(language.ErrorPattern) != 0 {
		patternForCompile =
			strings.ReplaceAll(language.ErrorPattern, "$message", template.Pattern)
	} else {
		patternForCompile = template.Pattern + "$stacktrace"
	}

	patternForCompile =
		strings.ReplaceAll(patternForCompile, "$stacktrace", defaultStackTraceRegex)

	if len(template.StackTracePattern) != 0 {
		var err error
		stackTracePattern, err = regexp.Compile(template.StackTracePattern)
		if err != nil {
			// TODO: should not panic!
			panic(err)
		}
	}

	compiledPattern, err := regexp.Compile("(?m)^" + patternForCompile + "$")
	if err != nil {
		// TODO: should not panic!
		panic(err)
	}

	*tmps = append(*tmps, &CompiledErrorTemplate{
		ErrorTemplate:     template,
		Language:          language,
		Pattern:           compiledPattern,
		StackTracePattern: stackTracePattern,
	})

	return (*tmps)[len(*tmps)-1]
}

func (tmps ErrorTemplates) Find(msg string) *CompiledErrorTemplate {
	for _, tmp := range tmps {
		if tmp.Pattern.MatchString(msg) {
			return tmp
		}
	}
	return nil
}

func (tmps ErrorTemplates) CompileAll() {
	for _, tmp := range tmps {
		if tmp.Language == nil {
			continue
		}
		tmp.Language.Compile()
	}
}
