package errgoengine

type MainError struct {
	ErrorNode *StackTraceEntry
	Document  *Document
	Nearest   SyntaxNode
	Context   any
}

func (err *MainError) DocumentPath() string {
	return err.ErrorNode.DocumentPath
}

type ContextData struct {
	*Store
	Analyzer            LanguageAnalyzer
	WorkingPath         string
	CurrentDocumentPath string
	Variables           map[string]string
	TraceStack          TraceStack
	MainError           *MainError
}

func NewContextData(store *Store, workingPath string) *ContextData {
	return &ContextData{
		Store:       store,
		WorkingPath: workingPath,
		Variables:   make(map[string]string),
		TraceStack:  TraceStack{},
	}
}

func (data *ContextData) MainDocumentPath() string {
	if data.MainError != nil && data.MainError.ErrorNode != nil {
		return data.MainError.DocumentPath()
	}
	return data.CurrentDocumentPath
}

func (data *ContextData) FindSymbol(name string, pos int) Symbol {
	path := data.MainDocumentPath()
	sym := data.Store.FindSymbol(path, name, pos)
	if sym == nil {
		return data.Analyzer.FindSymbol(name)
	}
	return sym
}

func (data *ContextData) AddVariable(name string, value string) {
	if data.Variables == nil {
		data.Variables = make(map[string]string)
	}

	data.Variables[name] = value
}

func (data *ContextData) AddVariables(vars map[string]string) {
	if len(vars) == 0 {
		return
	}

	if data.Variables == nil {
		data.Variables = make(map[string]string)
	}

	for k, v := range vars {
		data.Variables[k] = v
	}
}
