package java

import lib "github.com/nedpals/errgoengine"

var ClassCastException = lib.ErrorTemplate{
	Name:    "ClassCastException",
	Pattern: runtimeErrorPattern("java.lang.ClassCastException", `class (?P<currentClassName>\S+) cannot be cast to class (?P<targetClassName>\S+)`),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when there is an attempt to cast an object to a type that it is not compatible with.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		gen.Add("Use the correct type for casting", func(s *lib.BugFixSuggestion) {
			s.AddStep(
				"Ensure that the casting is done to the correct type. In this case, you should cast to `%s` instead of `%s`.",
				"String", cd.Variables["targetClassName"])
		})
	},
}
