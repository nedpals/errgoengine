package lib

import (
	sitter "github.com/smacker/go-tree-sitter"
)

type Node struct {
	*sitter.Node
	Doc  *Document
	text string
}

func (n Node) Text() string {
	return n.text
}

func (n Node) ChildByFieldName(field string) Node {
	cNode := n.Node.ChildByFieldName(field)
	return WrapNode(n.Doc, cNode)
}

func (n Node) Parent() Node {
	return WrapNode(n.Doc, n.Node.Parent())
}

func (n Node) NamedChild(idx int) Node {
	cNode := n.Node.NamedChild(idx)
	return WrapNode(n.Doc, cNode)
}

func (n Node) LastNamedChild() Node {
	len := n.Node.NamedChildCount()
	return n.NamedChild(int(len) - 1)
}

func (n Node) Child(idx int) Node {
	cNode := n.Node.Child(idx)
	return WrapNode(n.Doc, cNode)
}

func (n Node) StartPosition() Position {
	p := n.Node.StartPoint()
	return Position{
		Line:   int(p.Row),
		Column: int(p.Column),
		Index:  int(n.Node.StartByte()),
	}
}

func (n Node) EndPosition() Position {
	p := n.Node.EndPoint()
	return Position{
		Line:   int(p.Row),
		Column: int(p.Column),
		Index:  int(n.Node.StartByte()),
	}
}

func (n Node) Location() Location {
	return Location{
		DocumentPath: n.Doc.Path,
		Position:     n.StartPosition(),
	}
}

func (n Node) RawNode() *sitter.Node {
	return n.Node
}

func WrapNode(doc *Document, n *sitter.Node) Node {
	return Node{
		text: n.Content([]byte(doc.Contents)),
		Doc:  doc,
		Node: n,
	}
}
