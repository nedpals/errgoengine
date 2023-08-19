package lib

import (
	"regexp"
)

type ErrorTemplate struct {
	Name           string
	Pattern        string
	OnGenExplainFn GenExplainFn
	OnGenBugFixFn  GenBugFixFn
}

type compiledErrorTemplate struct {
	ErrorTemplate
	Language *Language
	Pattern  *regexp.Regexp
}

type ErrorTemplates []*compiledErrorTemplate

func (tmps *ErrorTemplates) Add(language *Language, template ErrorTemplate) *compiledErrorTemplate {
	patternForCompile := "(?m)^" + template.Pattern + `(?P<stacktrace>(?:.|\s)*)$`
	compiledPattern, err := regexp.Compile(patternForCompile)
	if err != nil {
		// TODO: should not panic!
		panic(err)
	}

	*tmps = append(*tmps, &compiledErrorTemplate{
		ErrorTemplate: template,
		Language:      language,
		Pattern:       compiledPattern,
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
