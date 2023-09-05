package errgoengine

import (
	"testing"

	testutils "github.com/nedpals/errgoengine/test_utils"
)

func TestDepGraph(t *testing.T) {
	graph := DepGraph{}

	// Adding
	graph.Add("a", map[string]string{"b": "c"})
	graph.Add("d", map[string]string{"e": "c"})

	testutils.Equals(t, graph.Has("a"), true)
	testutils.Equals(t, graph.Has("c"), true)

	testutils.Equals(t, graph["a"].Path, "a")
	testutils.EqualsMap(t, graph["a"].Graph, graph)
	testutils.EqualsMap(t, graph["a"].Dependencies, map[string]string{"b": "c"})

	testutils.Equals(t, graph["d"].Path, "d")
	testutils.EqualsMap(t, graph["d"].Graph, graph)
	testutils.EqualsMap(t, graph["d"].Dependencies, map[string]string{"e": "c"})

	testutils.EqualsList(t, graph["c"].DependentPaths(), []string{"a", "d"})

	// Deleting
	testutils.ExpectNoError(t, graph.Detach("d", "c"))
	testutils.Equals(t, graph.Has("d"), true)
	testutils.Equals(t, graph.Has("c"), true)
	testutils.EqualsList(t, graph["c"].DependentPaths(), []string{"a"})

	testutils.ExpectNoError(t, graph.Detach("a", "c"))
	testutils.Equals(t, graph.Has("a"), true)
	testutils.Equals(t, graph.Has("c"), false)
}
