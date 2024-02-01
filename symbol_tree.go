package errgoengine

type SymbolTree struct {
	Parent   *SymbolTree
	StartPos Position
	EndPos   Position
	Symbols  map[string]Symbol
	Scopes   []*SymbolTree
}

func (tree *SymbolTree) CreateChildFromNode(n SyntaxNode) *SymbolTree {
	nearest := tree.GetNearestScopedTree(n.StartPosition().Index)
	if nearest.StartPos.Eq(n.StartPosition()) && nearest.EndPos.Eq(n.EndPosition()) {
		return nearest
	}

	return &SymbolTree{
		Parent:   tree,
		StartPos: n.StartPosition(),
		EndPos:   n.EndPosition(),
		Symbols:  map[string]Symbol{},
	}
}

func (tree *SymbolTree) FindSymbolsByClause(findFn func(sym Symbol) bool) []Symbol {
	symbols := []Symbol{}

	for _, sym := range tree.Symbols {
		if findFn(sym) {
			symbols = append(symbols, sym)
		}
	}

	if tree.Parent != nil {
		symbols = append(symbols, tree.Parent.FindSymbolsByClause(findFn)...)
	}

	return symbols
}

func (tree *SymbolTree) Find(name string) Symbol {
	for _, sym := range tree.Symbols {
		if sym.Name() == name {
			return sym
		}
	}

	if tree.Parent != nil {
		return tree.Parent.Find(name)
	}

	return nil
}

func (tree *SymbolTree) GetNearestScopedTree(index int) *SymbolTree {
	if tree.Scopes != nil {
		for _, scopedTree := range tree.Scopes {
			if scopedTree == tree || scopedTree == nil {
				continue
			}

			if index >= scopedTree.StartPos.Index && index <= scopedTree.EndPos.Index {
				return scopedTree.GetNearestScopedTree(index)
			}
		}
	}
	return tree
}

func (tree *SymbolTree) GetSymbolByNode(node SyntaxNode) Symbol {
	// Get nearest tree
	nearestTree := tree.GetNearestScopedTree(node.StartPosition().Index)
	return nearestTree.Find(node.Text())
}

func (tree *SymbolTree) Add(sym Symbol) {
	if tree.Symbols == nil {
		tree.Symbols = make(map[string]Symbol)
	}

	tree.Symbols[sym.Name()] = sym
	loc := sym.Location()
	if loc.StartPos.Index < tree.StartPos.Index {
		tree.StartPos = loc.StartPos
	}

	if loc.EndPos.Index > tree.EndPos.Index {
		tree.EndPos = loc.EndPos
	}

	if cSym := CastChildrenSymbol(sym); cSym != nil && cSym.Children() != tree {
		if tree.Scopes == nil {
			tree.Scopes = []*SymbolTree{}
		}

		// check if already exists
		for _, scopedTree := range tree.Scopes {
			if scopedTree == cSym.Children() {
				return
			}
		}

		tree.Scopes = append(tree.Scopes, cSym.Children())
		cSym.Children().Parent = tree

		if cSym.Children().StartPos.Index < tree.StartPos.Index {
			tree.StartPos = cSym.Children().StartPos
		}

		if cSym.Children().EndPos.Index > tree.EndPos.Index {
			tree.EndPos = cSym.Children().EndPos
		}
	}
}
