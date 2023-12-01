package errgoengine

type Store struct {
	DepGraph  DepGraph
	Documents map[string]*Document
	Symbols   map[string]*SymbolTree
}

func NewEmptyStore() *Store {
	return &Store{
		DepGraph:  DepGraph{},
		Documents: map[string]*Document{},
		Symbols:   map[string]*SymbolTree{},
	}
}

func (store *Store) FindSymbol(docPath string, name string, pos int) Symbol {
	// Find local symbols first
	tree := store.Symbols[docPath]

	if pos != -1 {
		// go innerwards first
		for len(tree.Scopes) != 0 {
			found := false
			for _, s := range tree.Scopes {
				if pos >= s.StartPos.Index && pos <= s.EndPos.Index {
					found = true
					tree = s
					break
				}
			}
			if !found {
				break
			}
		}
	}

	if tree != nil {
		parent := tree

		// search innerwards first then outside
		for parent != nil {
			if sym := parent.Find(name); sym != nil {
				return sym
			} else {
				if parent == tree.Parent {
					break
				}
				parent = tree.Parent
			}
		}
	}

	return nil
}

func (store *ContextData) AddDocument(doc *Document) *Document {
	if store.Documents == nil {
		store.Documents = make(map[string]*Document)
	}

	store.Documents[doc.Path] = doc
	return doc
}

func (store *Store) InitOrGetSymbolTree(docPath string) *SymbolTree {
	if store.Symbols == nil {
		store.Symbols = make(map[string]*SymbolTree)
	}

	if store.Symbols[docPath] == nil {
		store.Symbols[docPath] = &SymbolTree{
			DocumentPath: docPath,
			Symbols:      make(map[string]Symbol),
		}
	}

	return store.Symbols[docPath]
}
