package lib

import (
	"regexp"
)

type ErrorTemplate struct {
	Name              string
	Pattern           string
	StackTracePattern string
	OnGenExplainFn    GenExplainFn
	OnGenBugFixFn     GenBugFixFn
}

type compiledErrorTemplate struct {
	ErrorTemplate
	Language          *Language
	Pattern           *regexp.Regexp
	StackTracePattern *regexp.Regexp
}

type ErrorTemplates []*compiledErrorTemplate

func (tmps *ErrorTemplates) Add(language *Language, template ErrorTemplate) *compiledErrorTemplate {
	var stackTracePattern *regexp.Regexp

	patternForCompile := template.Pattern
	if len(template.StackTracePattern) == 0 {
		patternForCompile += `(?P<stacktrace>(?:.|\s)*)`
	} else {
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

	*tmps = append(*tmps, &compiledErrorTemplate{
		ErrorTemplate:     template,
		Language:          language,
		Pattern:           compiledPattern,
		StackTracePattern: stackTracePattern,
	})

	return (*tmps)[len(*tmps)-1]
}

func (tmps ErrorTemplates) Find(msg string) *compiledErrorTemplate {
	for _, tmp := range tmps {
		if tmp.Pattern.MatchString(msg) {
			return tmp
		}
	}
	return nil
}
