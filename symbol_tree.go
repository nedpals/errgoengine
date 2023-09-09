package errgoengine

type SymbolTree struct {
	Parent       *SymbolTree
	StartPos     Position
	EndPos       Position
	DocumentPath string
	Symbols      map[string]Symbol
	Scopes       []*SymbolTree
}

func (tree *SymbolTree) CreateChildFromNode(n SyntaxNode) *SymbolTree {
	return &SymbolTree{
		Parent:       tree,
		StartPos:     n.StartPosition(),
		EndPos:       n.EndPosition(),
		DocumentPath: tree.DocumentPath,
		Symbols:      map[string]Symbol{},
	}
}

func (tree *SymbolTree) Find(name string) Symbol {
	for _, sym := range tree.Symbols {
		if sym.Name() == name {
			return sym
		}
	}
	return nil
}

func (tree *SymbolTree) Add(sym Symbol) {
	if tree.Symbols == nil {
		tree.Symbols = make(map[string]Symbol)
	}

	tree.Symbols[sym.Name()] = sym
	// TODO: create tree both in the parent and in the child symbol

	if sym.Location().Position.Index < tree.StartPos.Index {
		tree.StartPos = sym.Location().Position
	}

	if sym.Location().Index > tree.EndPos.Index {
		tree.EndPos = sym.Location().Position
	}

	if cSym := CastChildrenSymbol(sym); cSym != nil {
		tree.Scopes = append(tree.Scopes, cSym.Children())
		cSym.Children().Parent = tree
	}
}
