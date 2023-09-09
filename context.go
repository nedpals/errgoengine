package errgoengine

type MainError struct {
	ErrorNode *StackTraceEntry
	Document  *Document
	Nearest   Node
}

func (err MainError) DocumentPath() string {
	return err.ErrorNode.DocumentPath
}

type ContextData struct {
	*Store
	WorkingPath         string
	CurrentDocumentPath string
	Variables           map[string]string
	TraceStack          TraceStack
	MainError           MainError
}

func (data *ContextData) MainDocumentPath() string {
	if data.MainError.ErrorNode != nil {
		return data.MainError.DocumentPath()
	}
	return data.CurrentDocumentPath
}

func (data *ContextData) FindSymbol(name string, pos int) Symbol {
	path := data.MainDocumentPath()
	return data.Store.FindSymbol(path, name, pos)
}

func (data *ContextData) AnalyzeValue(n Node) Symbol {
	return n.Doc.Language.ValueAnalyzer(data, n)
}

func (data *ContextData) AddVariable(name string, value string) {
	if data.Variables == nil {
		data.Variables = make(map[string]string)
	}

	data.Variables[name] = value
}
