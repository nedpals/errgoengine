package errgoengine

type GenExplainFn func(*ContextData) string
type GenBugFixFn func(*ContextData) []BugFix

type BugFix struct {
	Content string // explanation
	Code    string
}
