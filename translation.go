package errgoengine

import (
	"fmt"
	"strings"
)

type GenExplainFn func(*ContextData, *ExplainGenerator)

type ExplainGenerator struct {
	errorName string
	mainExp   *strings.Builder
	sections  map[string]*ExplainGenerator
}

func (gen *ExplainGenerator) Add(text string, data ...any) {
	if gen.mainExp == nil {
		gen.mainExp = &strings.Builder{}
	}

	if len(data) != 0 {
		gen.mainExp.WriteString(fmt.Sprintf(text, data...))
	} else {
		gen.mainExp.WriteString(text)
	}
}

func (gen *ExplainGenerator) CreateSection(name string) *ExplainGenerator {
	if gen.sections == nil {
		gen.sections = map[string]*ExplainGenerator{}
	}
	_, ok := gen.sections[name]
	if !ok {
		gen.sections[name] = &ExplainGenerator{}
	}
	return gen.sections[name]
}

type GenBugFixFn func(*ContextData, *BugFixGenerator)

type BugFixSuggestion struct {
	Title       string
	Description string
	Steps       []*BugFixStep
}

func (gen *BugFixSuggestion) AddStep(content string, d ...any) *BugFixStep {
	if gen.Steps == nil {
		gen.Steps = []*BugFixStep{}
	}

	gen.Steps = append(gen.Steps, &BugFixStep{
		Content: fmt.Sprintf(content, d...),
	})

	return gen.Steps[len(gen.Steps)-1]
}

func (gen *BugFixSuggestion) AddDescription(exp string, d ...any) {
	gen.Description = fmt.Sprintf(exp, d...)
}

type BugFixStep struct {
	Content string
	Fixes   []SuggestedFix
}

func (step *BugFixStep) AddFix(fix SuggestedFix) *BugFixStep {
	if step.Fixes == nil {
		step.Fixes = []SuggestedFix{}
	}

	step.Fixes = append(step.Fixes, fix)
	return step
}

type SuggestedFix struct {
	StartPosition Position
	EndPosition   Position
	NewText       string
	Description   string
}

type BugFixGenerator struct {
	Suggestions []*BugFixSuggestion
}

func (gen *BugFixGenerator) Add(title string, makerFn func(s *BugFixSuggestion)) {
	if gen.Suggestions == nil {
		gen.Suggestions = []*BugFixSuggestion{}
	}

	suggestion := &BugFixSuggestion{Title: title}
	makerFn(suggestion)

	gen.Suggestions = append(gen.Suggestions, suggestion)
}
