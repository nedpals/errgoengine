package java

import (
	lib "github.com/nedpals/errgoengine"
	sitter "github.com/smacker/go-tree-sitter"
)

type characterExpectedFixKind int

const (
	characterExpectedFixUnknown      characterExpectedFixKind = 0
	characterExpectedFixWrapFunction characterExpectedFixKind = iota
)

type characterInsertDirection int

const (
	characterInsertDirectionLeft characterInsertDirection = iota
	characterInsertDirectionRight
)

type characterExpectedErrorCtx struct {
	fixKind   characterExpectedFixKind
	direction characterInsertDirection
}

var CharacterExpectedError = lib.ErrorTemplate{
	Name:              "CharacterExpectedError",
	Pattern:           comptimeErrorPattern(`'(?P<character>\S+)'(?: or '(?P<altCharacter>\S+)')? expected`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		iCtx := characterExpectedErrorCtx{
			direction: characterInsertDirectionLeft,
		}

		// TODO: check if node is parsable
		rootNode := m.Document.Tree.RootNode()
		cursor := sitter.NewTreeCursor(rootNode)
		rawNearestMissingNode := nearestMissingNodeFromPos(cursor, m.ErrorNode.StartPos)
		if rawNearestMissingNode == nil {
			if rawNearestMissingNode2 := nearestNodeFromPos2(cursor, m.ErrorNode.StartPos); rawNearestMissingNode2 != nil {
				rawNearestMissingNode = rawNearestMissingNode2
			}
		}

		if rawNearestMissingNode != nil {
			if rawNearestMissingNode.IsExtra() {
				// go back
				rawNearestMissingNode = rawNearestMissingNode.PrevSibling()
			}

			if rawNearestMissingNode.IsNamed() {
				iCtx.direction = characterInsertDirectionRight
			}

			nearestMissingNode := lib.WrapNode(m.Document, rawNearestMissingNode)
			m.Nearest = nearestMissingNode
		}

		m.Context = iCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when there is an unexpected character in the code, and '%s' is expected.", cd.Variables["character"])

		// ctx := cd.MainError.Context.(CharacterExpectedErrorCtx)

		// switch ctx.kind {
		// case cannotBeAppliedMismatchedArgCount:
		// 	gen.Add("This error occurs when there is an attempt to apply a method with an incorrect number of arguments.")
		// case cannotBeAppliedMismatchedArgType:
		// 	gen.Add("This error occurs when there is an attempt to apply a method with arguments that do not match the method signature.")
		// default:
		// 	gen.Add("unable to determine.")
		// }
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(characterExpectedErrorCtx)

		gen.Add("Add the missing character", func(s *lib.BugFixSuggestion) {
			step := s.AddStep("Ensure that the array declaration has the correct syntax by adding the missing `%s`.", cd.Variables["character"])
			if ctx.direction == characterInsertDirectionLeft {
				step.AddFix(lib.FixSuggestion{
					NewText:       cd.Variables["character"],
					StartPosition: cd.MainError.Nearest.StartPosition(),
					EndPosition:   cd.MainError.Nearest.StartPosition(),
				})
			} else if ctx.direction == characterInsertDirectionRight {
				step.AddFix(lib.FixSuggestion{
					NewText:       cd.Variables["character"],
					StartPosition: cd.MainError.Nearest.EndPosition(),
					EndPosition:   cd.MainError.Nearest.EndPosition(),
				})
			}
		})

		// switch ctx.fixKind {
		// case characterExpectedFixWrapFunction:
		// 	gen.Add("Use the correct syntax", func(s *lib.BugFixSuggestion) {
		// 		startPos := cd.MainError.Nearest.StartPosition()
		// 		space := getSpaceFromBeginning(cd.MainError.Document, startPos.Line, startPos.Column)

		// 		s.AddStep("Use a valid statement or expression within a method or block.").
		// 			AddFix(lib.FixSuggestion{
		// 				NewText: space + "public void someMethod() {\n" + space,
		// 				StartPosition: lib.Position{
		// 					Line: cd.MainError.Nearest.StartPosition().Line,
		// 				},
		// 				EndPosition: lib.Position{
		// 					Line: cd.MainError.Nearest.StartPosition().Line,
		// 				},
		// 			}).
		// 			AddFix(lib.FixSuggestion{
		// 				NewText:       "\n" + space + "}",
		// 				StartPosition: cd.MainError.Nearest.EndPosition(),
		// 				EndPosition:   cd.MainError.Nearest.EndPosition(),
		// 			})
		// 	})
		// }
	},
}

func nearestNodeFromPos2(cursor *sitter.TreeCursor, pos lib.Position) *sitter.Node {
	defer cursor.GoToParent()

	// hope it executes to avoid stack overflow
	if !cursor.GoToFirstChild() {
		return nil
	}

	var nearest *sitter.Node

	for {
		currentNode := cursor.CurrentNode()
		pointA := currentNode.StartPoint()
		if pointA.Row+1 > uint32(pos.Line) {
			break
		}

		if (nearest == nil && int(pointA.Row+1) <= pos.Line) || currentNode.StartByte() >= nearest.StartByte() {
			nearest = currentNode
		}

		if currentNode.ChildCount() != 0 {
			if nearestFromInner := nearestNodeFromPos2(cursor, pos); nearestFromInner != nil {
				if (nearest == nil && int(nearestFromInner.StartPoint().Row+1) <= pos.Line) || nearestFromInner.StartByte() >= nearest.StartByte() {
					nearest = nearestFromInner
				}
			}
		}

		if !cursor.GoToNextSibling() {
			break
		}
	}

	return nearest
}
