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

type BugFixGenerator struct {
	Steps []*BugFixStep
}

func (gen *BugFixGenerator) AddStep(exp string, d ...any) *BugFixStep {
	if gen.Steps == nil {
		gen.Steps = []*BugFixStep{}
	}

	gen.Steps = append(gen.Steps, &BugFixStep{
		Explanation: fmt.Sprintf(exp, d...),
	})

	return gen.Steps[len(gen.Steps)-1]
}

// TODO:
func (gen *BugFixGenerator) LookFor(nodeType string) {

}

type BugFixStep struct {
	Explanation string
	Fixes       []SuggestedFix
}

func (step *BugFixStep) AddFix(newText string, pos Position, replace bool) *BugFixStep {
	if step.Fixes == nil {
		step.Fixes = []SuggestedFix{}
	}

	step.Fixes = append(step.Fixes, SuggestedFix{
		Position: pos,
		Replace:  replace,
		NewText:  newText,
	})

	return step
}

type SuggestedFix struct {
	Position Position
	Replace  bool
	NewText  string
}
