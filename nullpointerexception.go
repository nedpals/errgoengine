package main

var NullPointerException = ErrorTemplate{
	Name:    "NullPointerException",
	Pattern: `Exception in thread "(?P<thread>\w+)" java\.lang\.NullPointerException`,
	OnGenExplainFn: func(cd *ContextData) string {
		// sb := &strings.Builder{}
		isSystemOut := false

		if cd.MainError.Nearest.Type() == "expression_statement" {
			exprNode := cd.MainError.Nearest.Child(0)

			// NOTE: i hope we will be able to parse the entire
			// java system library without hardcoding the definitions lol
			if exprNode.Type() == "method_invocation" {
				objNode := exprNode.ChildByFieldName("object")
				nameNode := exprNode.ChildByFieldName("name")

				if objNode.Text() == "System.out" && nameNode.Text() == "println" {
					isSystemOut = true
				}
			}

			if isSystemOut {

			}
		}

		// switch cd.MainError.Nearest.Type() {
		// case "method_invocation":

		// }

		// sb.WriteString()
		// return "Your program try to access or manipulate an object reference that is currently pointing to `null`, meaning it doesn't refer to any actual object in memory. This typically happens when you forget to initialize an object before using it, or when you try to access an object that hasn't been properly assigned a value. "
		return cd.MainError.Nearest.Node.String()
	},
	OnGenBugFixFn: func(cd *ContextData) []BugFix {
		return []BugFix{}
	},
}
