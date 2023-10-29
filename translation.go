package errgoengine

import (
	"fmt"
	"strings"
)

type GenExplainFn func(*ContextData, *ExplainGenerator)

type ExplainGenerator struct {
	mainExp  *strings.Builder
	sections map[string]*ExplainGenerator
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
	Fixes []BugFix
}

// TODO:
func (gen *BugFixGenerator) AddStep() {
	if gen.Fixes == nil {
		gen.Fixes = []BugFix{}
	}

}

// TODO:
func (gen *BugFixGenerator) LookFor(nodeType string) {

}

type BugFix struct {
	Position    Position
	Node        SyntaxNode
	Explanation string
	Fixes       []SuggestedFix
}

type SuggestedFix struct {
	Position Position
	NewText  string
}
