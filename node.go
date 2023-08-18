package main

import sitter "github.com/smacker/go-tree-sitter"

type Position struct {
	Line   int
	Column int
	Index  int
}

type Location struct {
	DocumentPath string
	Position
}

type Node struct {
	*sitter.Node
	doc  *Document
	text string
}

func (n Node) Text() string {
	return n.text
}

func (n Node) ChildByFieldName(field string) Node {
	cNode := n.Node.ChildByFieldName(field)
	return wrapNode(n.doc, cNode)
}

func (n Node) Parent() Node {
	return wrapNode(n.doc, n.Node.Parent())
}

func (n Node) NamedChild(idx int) Node {
	cNode := n.Node.NamedChild(idx)
	return wrapNode(n.doc, cNode)
}

func (n Node) LastNamedChild() Node {
	len := n.Node.NamedChildCount()
	return n.NamedChild(int(len) - 1)
}

func (n Node) Child(idx int) Node {
	cNode := n.Node.Child(idx)
	return wrapNode(n.doc, cNode)
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
		DocumentPath: n.doc.Path,
		Position:     n.StartPosition(),
	}
}

func wrapNode(doc *Document, n *sitter.Node) Node {
	return Node{
		text: n.Content([]byte(doc.Contents)),
		doc:  doc,
		Node: n,
	}
}
