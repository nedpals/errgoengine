package lib

import (
	"fmt"
	"regexp"
	"strings"
)

var SymPrefix = "%d.sym.%d"

type ISymbolCapture interface {
	Compile(prefix, tag string, sb *strings.Builder)
	SymKind() SymbolKind
}

type ISymbolCaptureList []ISymbolCapture

func (list ISymbolCaptureList) Compile(prefix, tag string, sb *strings.Builder) {
	sb.WriteString("[")
	for idx, sc := range list {
		sc.Compile(fmt.Sprintf(SymPrefix, idx, sc.SymKind()), "", sb)
	}
	sb.WriteString("]+")
	if len(tag) != 0 {
		sb.WriteString(" @")
		if len(prefix) != 0 {
			sb.WriteString(prefix)
			sb.WriteByte('.')
		}
		sb.WriteString(tag)
	}
}

func (list ISymbolCaptureList) SymKind() SymbolKind {
	return SymbolKindUnknown
}

// ^\d+.sym.\d+
var SymPrefixRegex = regexp.MustCompile(fmt.Sprintf("^%s", strings.ReplaceAll(SymPrefix, "%d", `\d+`)))

type SymbolCapture struct {
	Query    string
	Kind     SymbolKind
	Field    string
	Optional bool

	NameNode       *SymbolCapture
	ParameterNodes *SymbolCapture
	ReturnTypeNode *SymbolCapture // should be nil for top-level symbols
	ContentNode    *SymbolCapture
	BodyNode       *SymbolCapture

	Children []*SymbolCapture
}

func (cap SymbolCapture) SymKind() SymbolKind {
	return cap.Kind
}

func (cap SymbolCapture) Compile(prefix, tag string, sb *strings.Builder) {
	isAlternations := strings.HasPrefix(cap.Query, "[") && strings.HasSuffix(cap.Query, "]")
	isSingle := len(cap.Children) < 2
	parCount := countSuffix(cap.Query, ')')

	if len(cap.Field) != 0 {
		sb.WriteString(cap.Field)
		sb.WriteString(": ")
	}

	if !isAlternations {
		sb.WriteByte('(')
	}

	if parCount > 0 {
		sb.WriteString(cap.Query[:len(cap.Query)-parCount])
		sb.WriteByte(' ')
	} else {
		sb.WriteString(cap.Query)
	}

	if len(cap.Children) != 0 {
		if !isSingle {
			sb.WriteByte('\n')
			sb.WriteByte('[')
		}

		for i, c := range cap.Children {
			sb.WriteByte('\n')
			c.Compile(
				fmt.Sprintf(
					"%s.child.%d",
					prefix,
					i,
				),
				"",
				sb,
			)
		}

		if !isSingle {
			sb.WriteString("\n]*")

			if len(tag) != 0 {
				sb.WriteString(" @" + prefix + "." + tag)
			}
		}
	} else {
		if cap.ReturnTypeNode != nil {
			sb.WriteRune('\n')
			cap.ReturnTypeNode.Compile(prefix, "return-type", sb)
		}

		if cap.NameNode != nil {
			sb.WriteRune('\n')
			cap.NameNode.Compile(prefix, "name", sb)
		}

		if cap.ParameterNodes != nil {
			sb.WriteRune('\n')
			cap.ParameterNodes.Compile(prefix, "parameters", sb)
		}

		if cap.ContentNode != nil {
			sb.WriteRune('\n')
			cap.ContentNode.Compile(prefix, "content", sb)
		}

		if cap.BodyNode != nil {
			sb.WriteByte('\n')
			sb.WriteString("body: (_) @")
			if len(prefix) != 0 {
				sb.WriteString(prefix)
				sb.WriteByte('.')
			}
			sb.WriteString("body")
			// cap.BodyNode.Compile(prefix, "body", sb)
		}
	}

	if parCount > 0 {
		sb.WriteString(strings.Repeat(")", parCount))
	}

	if !isAlternations {
		sb.WriteByte(')')
	}

	if cap.Optional {
		sb.WriteByte('?')
	}

	if isSingle && len(tag) != 0 {
		sb.WriteString(" @")

		if len(prefix) != 0 {
			sb.WriteString(prefix)
			sb.WriteByte('.')
		}

		sb.WriteString(tag)
	}
}

func SymCaptureToListPtr(list []*SymbolCapture) ISymbolCaptureList {
	captures := make(ISymbolCaptureList, len(list))
	for i, sc := range list {
		captures[i] = sc
	}
	return captures
}

func SymCaptureToList(list []SymbolCapture) ISymbolCaptureList {
	captures := make(ISymbolCaptureList, len(list))
	for i, sc := range list {
		captures[i] = sc
	}
	return captures
}

func countSuffix(str string, s byte) int {
	c := 0
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] == s {
			c++
		} else {
			return c
		}
	}
	return 0
}
