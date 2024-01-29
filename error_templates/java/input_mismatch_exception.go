package java

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
	sitter "github.com/smacker/go-tree-sitter"
)

var scannerMethodsToTypes = map[string]string{
	"nextInt":    "integer",
	"nextLong":   "long",
	"nextByte":   "byte",
	"nextChar":   "character",
	"nextDouble": "double",
	"nextFloat":  "float",
	"nextShort":  "short",
}

var InputMismatchException = lib.ErrorTemplate{
	Name:    "InputMismatchException",
	Pattern: runtimeErrorPattern("java.util.InputMismatchException", ""),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		query := "(method_invocation object: (_) name: (identifier) @fn-name arguments: (argument_list) (#eq? @fn-name \"nextInt\"))"
		for q := m.Nearest.Query(query); q.Next(); {
			if q.CurrentTagName() != "fn-name" {
				continue
			}

			node := q.CurrentNode()
			m.Nearest = node
		}
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		methodName := cd.MainError.Nearest.Text()
		expectedType := scannerMethodsToTypes[methodName]
		gen.Add(
			"This error occurs when a non-%s input is passed to the `%s()` method of the `Scanner` class.",
			expectedType,
			methodName)
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// get nearest block from position
		cursor := sitter.NewTreeCursor(cd.MainError.Document.RootNode().Node)
		rawNearestBlock := nearestNodeFromPosByType(cursor, "block", cd.MainError.Nearest.StartPosition())

		if rawNearestBlock != nil && !rawNearestBlock.IsNull() {
			nearestBlock := lib.WrapNode(cd.MainError.Document, rawNearestBlock)

			gen.Add("Add a try-catch for error handling", func(s *lib.BugFixSuggestion) {
				step := s.AddStep("Implement error handling to account for input mismatches and prompt the user for valid input.")

				wrapStatement(
					step,
					"try {",
					"} catch (InputMismatchException e) {\n\t<i>System.out.println(\"Invalid input. Please try again.\");\n\t}",
					lib.Location{
						StartPos: lib.Position{
							Line: nearestBlock.FirstNamedChild().StartPosition().Line,
						},
						EndPos: nearestBlock.LastNamedChild().EndPosition(),
					},
					true,
				)
			})
		}
	},
}

func nearestNodeFromPosByType(cursor *sitter.TreeCursor, expType string, pos lib.Position) *sitter.Node {
	cursor.GoToFirstChild()
	defer cursor.GoToParent()

	var nearest *sitter.Node

	for {
		currentNode := cursor.CurrentNode()
		pointA := currentNode.StartPoint()
		pointB := currentNode.EndPoint()

		fmt.Println(currentNode.Type(), currentNode.StartByte())

		// stop if node is above the position
		if uint32(pos.Line) < pointA.Row+1 {
			break
		}

		// check if the current node is the nearest
		if currentNode.Type() == expType && uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
			nearest = currentNode
		}

		if currentNode.ChildCount() != 0 && uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
			nearestFromInner := nearestNodeFromPosByType(cursor, expType, pos)
			// check if the nearest from the inner nodes is nearer than the current nearest
			if nearestFromInner != nil && !nearestFromInner.IsNull() {
				nearest = nearestFromInner
			}
		}

		if !cursor.GoToNextSibling() {
			break
		}

		fmt.Println("next")
	}

	return nearest
}
