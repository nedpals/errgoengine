package errgoengine

type SymbolTree struct {
	Parent       *SymbolTree
	StartPos     Position
	EndPos       Position
	DocumentPath string
	Symbols      map[string]Symbol
	scopeIdxs    map[*SymbolTree]int
	Scopes       []*SymbolTree
}

func (tree *SymbolTree) CreateChildFromNode(n SyntaxNode) *SymbolTree {
	nearest := tree.GetNearestScopedTree(n.StartPosition().Index)
	if nearest.StartPos.Eq(n.StartPosition()) && nearest.EndPos.Eq(n.EndPosition()) {
		return nearest
	}

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

	if tree.Parent != nil {
		return tree.Parent.Find(name)
	}

	return nil
}

func (tree *SymbolTree) GetNearestScopedTree(index int) *SymbolTree {
	if tree.Scopes != nil {
		for _, scopedTree := range tree.Scopes {
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

	if cSym := CastChildrenSymbol(sym); cSym != nil {
		if tree.Scopes == nil {
			tree.Scopes = []*SymbolTree{}
		}

		if tree.scopeIdxs == nil {
			tree.scopeIdxs = make(map[*SymbolTree]int)
		}

		if _, ok := tree.scopeIdxs[cSym.Children()]; ok {
			return
		}

		tree.Scopes = append(tree.Scopes, cSym.Children())
		tree.scopeIdxs[cSym.Children()] = len(tree.Scopes) - 1
		cSym.Children().Parent = tree

		if cSym.Children().StartPos.Index < tree.StartPos.Index {
			tree.StartPos = cSym.Children().StartPos
		}

		if cSym.Children().EndPos.Index > tree.EndPos.Index {
			tree.EndPos = cSym.Children().EndPos
		}
	}
}
