package java

import (
	"context"
	"strings"

	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/utils/levenshtein"
	sitter "github.com/smacker/go-tree-sitter"
)

type identifiedExpectedReasonKind int

const (
	identifierExpectedReasonUnknown            identifiedExpectedReasonKind = 0
	identifierExpectedReasonClassInterfaceEnum identifiedExpectedReasonKind = iota
)

type identifierExpectedFixKind int

const (
	identifierExpectedFixUnknown      identifierExpectedFixKind = 0
	identifierExpectedFixWrapFunction identifierExpectedFixKind = iota
	identifierExpectedCorrectTypo     identifierExpectedFixKind = iota
)

type identifierExpectedErrorCtx struct {
	reasonKind  identifiedExpectedReasonKind
	typoWord    string // for identifierExpectedCorrectTypo. the word that is a typo
	wordForTypo string // for identifierExpectedCorrectTypo. the closest word to replace the typo
	fixKind     identifierExpectedFixKind
}

var IdentifierExpectedError = lib.ErrorTemplate{
	Name:              "IdentifierExpectedError",
	Pattern:           comptimeErrorPattern(`(?P<reason>\<identifier\>|class, interface, or enum) expected`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		iCtx := identifierExpectedErrorCtx{}

		// identify the reason
		switch cd.Variables["reason"] {
		case "class, interface, or enum":
			iCtx.reasonKind = identifierExpectedReasonClassInterfaceEnum
		default:
			iCtx.reasonKind = identifierExpectedReasonUnknown
		}

		// TODO: check if node is parsable
		if iCtx.reasonKind == identifierExpectedReasonClassInterfaceEnum {
			// use levenstein distance to check if the word is a typo
			tokens := []string{"class", "interface", "enum"}

			// get the nearest word
			nearestWord := ""
			wordToReplace := ""

			// get the contents of that line
			line := m.Document.LineAt(m.Nearest.StartPosition().Line)
			lineTokens := strings.Split(line, " ")

			// position
			nearestCol := 0

			for _, token := range tokens {
				for ltIdx, lineToken := range lineTokens {
					if levenshtein.ComputeDistance(token, lineToken) <= 3 {
						wordToReplace = lineToken
						nearestWord = token

						// compute the position of the word
						for i := 0; i < ltIdx; i++ {
							nearestCol += len(lineTokens[i]) + 1
						}

						// add 1 to nearestCol to get the portion of the word
						nearestCol++
						break
					}
				}
			}

			if nearestWord != "" {
				iCtx.wordForTypo = nearestWord
				iCtx.typoWord = wordToReplace
				iCtx.fixKind = identifierExpectedCorrectTypo

				targetPos := lib.Position{
					Line:   m.Nearest.StartPosition().Line,
					Column: nearestCol,
				}

				// get the nearest node of the word from the position
				initialNearest := m.Document.RootNode().NamedDescendantForPointRange(lib.Location{
					StartPos: targetPos,
					EndPos:   targetPos,
				})

				rawNearestNode := nearestNodeFromPos2(initialNearest.TreeCursor(), targetPos)
				if rawNearestNode != nil {
					m.Nearest = lib.WrapNode(m.Document, rawNearestNode)
				} else {
					m.Nearest = initialNearest
				}
			}
		} else if tree, err := sitter.ParseCtx(
			context.Background(),
			[]byte(m.Nearest.Text()),
			m.Document.Language.SitterLanguage,
		); err == nil && !tree.IsError() {
			iCtx.fixKind = identifierExpectedFixWrapFunction
		}

		m.Context = iCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		iCtx := cd.MainError.Context.(identifierExpectedErrorCtx)

		switch iCtx.reasonKind {
		case identifierExpectedReasonClassInterfaceEnum:
			gen.Add("This error occurs when there's a typo or the keyword `class`, `interface`, or `enum` is missing.")
		default:
			gen.Add("This error occurs when an identifier is expected, but an expression is found in a location where a statement or declaration is expected.")
		}

	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(identifierExpectedErrorCtx)

		switch ctx.fixKind {
		case identifierExpectedFixWrapFunction:
			gen.Add("Use the correct syntax", func(s *lib.BugFixSuggestion) {
				startPos := cd.MainError.Nearest.StartPosition()
				space := getSpaceFromBeginning(cd.MainError.Document, startPos.Line, startPos.Column)

				s.AddStep("Use a valid statement or expression within a method or block.").
					AddFix(lib.FixSuggestion{
						NewText: space + "public void someMethod() {\n" + space,
						StartPosition: lib.Position{
							Line: cd.MainError.Nearest.StartPosition().Line,
						},
						EndPosition: lib.Position{
							Line: cd.MainError.Nearest.StartPosition().Line,
						},
					}).
					AddFix(lib.FixSuggestion{
						NewText:       "\n" + space + "}",
						StartPosition: cd.MainError.Nearest.EndPosition(),
						EndPosition:   cd.MainError.Nearest.EndPosition(),
					})
			})
		case identifierExpectedCorrectTypo:
			gen.Add("Correct the typo", func(s *lib.BugFixSuggestion) {
				s.AddStep("Change `%s` to `%s` to properly declare the %s.", ctx.typoWord, ctx.wordForTypo, ctx.wordForTypo).
					AddFix(lib.FixSuggestion{
						NewText:       ctx.wordForTypo,
						StartPosition: cd.MainError.Nearest.StartPosition(),
						EndPosition:   cd.MainError.Nearest.EndPosition(),
					})
			})
		}
	},
}
