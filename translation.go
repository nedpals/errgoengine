package errgoengine

import (
	"fmt"
	"strings"
	"unicode"
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
	Title        string
	Description  string
	Steps        []*BugFixStep
	diffPosition Position
	Doc          *EditableDocument
}

func (gen *BugFixSuggestion) addStep(doc *EditableDocument, content string, d ...any) *BugFixStep {
	if gen.Steps == nil {
		gen.Steps = []*BugFixStep{}
	}

	if !strings.HasSuffix(content, ".") && (len(content) > 0 && !unicode.IsPunct(rune(content[len(content)-1]))) {
		content += "."
	}

	gen.Steps = append(gen.Steps, &BugFixStep{
		suggestion: gen,
		Content:    fmt.Sprintf(content, d...),
		Doc:        gen.Doc,
	})

	return gen.Steps[len(gen.Steps)-1]
}

func (gen *BugFixSuggestion) AddStep(content string, d ...any) *BugFixStep {
	s := gen.addStep(gen.Doc.Copy(), content, d...)
	s.isCopyable = true
	return s
}

func (gen *BugFixSuggestion) AddDescription(exp string, d ...any) {
	gen.Description = fmt.Sprintf(exp, d...)
}

type BugFixStep struct {
	suggestion    *BugFixSuggestion
	isCopyable    bool
	isSet         bool
	Doc           *EditableDocument
	StartLine     int
	AfterLine     int
	OrigStartLine int
	OrigAfterLine int
	Content       string
	DiffPosition  Position
	Fixes         []FixSuggestion
}

func (step *BugFixStep) AddFix(fix FixSuggestion) *BugFixStep {
	if step.Fixes == nil {
		step.Fixes = []FixSuggestion{}
	}

	step.Fixes = append(step.Fixes, fix)

	if !step.isSet {
		// get the start and end line after applying the diff
		step.StartLine = step.Fixes[0].StartPosition.Line
		step.AfterLine = step.Fixes[0].EndPosition.Line

		// get the original start and end line
		step.OrigStartLine = step.Fixes[0].StartPosition.Line
		step.OrigAfterLine = step.Fixes[0].EndPosition.Line

		if !step.isCopyable {
			// set diff position
			step.DiffPosition = Position{
				Line:   step.suggestion.diffPosition.Line,
				Column: step.suggestion.diffPosition.Column,
				Index:  step.suggestion.diffPosition.Index,
			}
		}

		step.isSet = true
	}

	fIdx := len(step.Fixes) - 1
	changeset := Changeset{
		NewText:  fix.NewText,
		StartPos: fix.StartPosition,
		EndPos:   fix.EndPosition,
	}

	// do not adjust position if the current fix is above the previous fix position
	// if fIdx >= 0 && step.Fixes[fIdx-1].StartPosition.Line <= fix.StartPosition.Line {
	if fIdx-1 >= 0 {
		changeset = changeset.Add(step.DiffPosition)
	}

	step.DiffPosition = step.DiffPosition.AddUnsafe(step.Doc.Apply(changeset))

	// change origStartLine only if
	// - the fix is a "deletion" and less than the current origStartLine
	// - the fix is an "insertion" or "replacement" and greater than the current origStartLine
	origStartLine2 := min(step.OrigStartLine, fix.StartPosition.Line)
	if len(fix.NewText) == 0 || fix.StartPosition.Line > step.OrigStartLine {
		step.OrigStartLine = origStartLine2
	}

	step.OrigAfterLine = max(step.OrigAfterLine, fix.EndPosition.Line)
	step.StartLine = min(step.StartLine, fix.StartPosition.Line+step.DiffPosition.Line)

	// if the diff position is negative, we need to set the after line to the latest position
	if step.DiffPosition.Line < 0 {
		step.AfterLine = fix.EndPosition.Line + step.DiffPosition.Line
	} else {
		step.AfterLine = max(step.AfterLine, fix.EndPosition.Line+step.DiffPosition.Line)
	}

	if !step.isCopyable {
		// set diff position
		step.suggestion.diffPosition = Position{
			Line:   step.DiffPosition.Line,
			Column: step.DiffPosition.Column,
			Index:  step.DiffPosition.Index,
		}
	}

	return step
}

type FixSuggestion struct {
	StartPosition Position
	EndPosition   Position
	NewText       string
	Description   string
}

type BugFixGenerator struct {
	Document    *Document
	Suggestions []*BugFixSuggestion
}

func (gen *BugFixGenerator) Add(title string, makerFn func(s *BugFixSuggestion)) {
	if gen.Suggestions == nil {
		gen.Suggestions = []*BugFixSuggestion{}
	}

	suggestion := &BugFixSuggestion{
		Title: title,
		Doc:   gen.Document.Editable(),
	}
	makerFn(suggestion)

	gen.Suggestions = append(gen.Suggestions, suggestion)
}
