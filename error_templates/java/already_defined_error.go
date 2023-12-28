package java

import (
	"fmt"
	"regexp"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type alreadyDefinedErrorCtx struct {
	NearestClass  lib.SyntaxNode
	NearestMethod lib.SyntaxNode
}

var AlreadyDefinedError = lib.ErrorTemplate{
	Name:              "AlreadyDefinedError",
	Pattern:           comptimeErrorPattern(`variable (?P<variable>\S+) is already defined in method (?P<symbolSignature>.+)`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		aCtx := alreadyDefinedErrorCtx{}
		rootNode := m.Document.RootNode()
		rawQuery := parseSymbolSignature(cd.Variables["symbolSignature"])
		pos := m.ErrorNode.StartPos

		// get the nearest class declaration first based on error location
		lib.QueryNode(rootNode, strings.NewReader("(class_declaration) @class"), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				pointA := c.Node.StartPoint()
				pointB := c.Node.EndPoint()
				if uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
					node := lib.WrapNode(m.Nearest.Doc, c.Node)
					aCtx.NearestClass = node
					return false
				}
			}
			return true
		})

		// get the nearest method declaration based on symbol signature
		lib.QueryNode(aCtx.NearestClass, strings.NewReader(rawQuery), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				aCtx.NearestMethod = node
				return false
			}
			return true
		})

		m.Context = aCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when you try to declare a variable with a name that is already in use within the same scope.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		gen.Add("Remove redeclaration", func(s *lib.BugFixSuggestion) {
			s.AddStep("To resolve the already defined error, remove the attempt to redeclare the variable '%s'.", cd.Variables["variable"]).
				AddFix(lib.FixSuggestion{
					NewText:       "",
					StartPosition: lib.Position{Line: cd.MainError.Nearest.StartPosition().Line, Column: 0},
					EndPosition:   cd.MainError.Nearest.EndPosition(),
					Description:   fmt.Sprintf("Since '%s' is already declared earlier in the method, you don't need to declare it again.", cd.Variables["variable"]),
				})
		})

		gen.Add("Assign a new value", func(s *lib.BugFixSuggestion) {
			dupeVarType := cd.MainError.Nearest.ChildByFieldName("type")
			dupeVarDeclarator := cd.MainError.Nearest.ChildByFieldName("declarator")

			s.AddStep("If you intended to change the value of '%s', you can simply assign a new value to the existing variable.", cd.Variables["variable"]).
				AddFix(lib.FixSuggestion{
					NewText:       "",
					StartPosition: dupeVarType.StartPosition(),
					EndPosition:   dupeVarDeclarator.StartPosition(),
					Description:   fmt.Sprintf("This way, you update the value of '%s' without redeclaring it.", cd.Variables["variable"]),
				})
		})
	},
}

var symbolSigRegex = regexp.MustCompile(`^(?m)(\S+)\((.+)\)$`)

// converts the signature into a tree-sitter query
func parseSymbolSignature(str string) string {
	sb := &strings.Builder{}
	methodName := ""
	paramTypes := []string{}

	for _, submatches := range symbolSigRegex.FindAllStringSubmatch(str, -1) {
		for i, matchedContent := range submatches {
			switch i {
			case 1:
				methodName = matchedContent
			case 2:
				paramTypes = strings.Split(matchedContent, ",")
			}
		}
	}

	sb.WriteByte('(')
	sb.WriteString("(method_declaration name: (identifier) @method-name parameters: (formal_parameters")
	for i := range paramTypes {
		sb.WriteString(fmt.Sprintf(" (formal_parameter type: (_) @param-%d-type)", i))
	}
	sb.WriteString(")) @method")
	sb.WriteString(" (#eq? @method-name \"" + methodName + "\")")
	for i, expType := range paramTypes {
		sb.WriteString(fmt.Sprintf(" (#eq? @param-%d-type \"%s\")", i, expType))
	}
	sb.WriteByte(')')
	return sb.String()
}
