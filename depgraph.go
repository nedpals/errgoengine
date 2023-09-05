package errgoengine

import "fmt"

type DepGraph map[string]*DepNode

type DepNode struct {
	Graph        DepGraph
	Path         string            // path where the module/library/package is located
	Dependencies map[string]string // mapped as map[label]depPath
}

func (node *DepNode) GetDependencies() []*DepNode {
	deps := make([]*DepNode, len(node.Dependencies))

	i := 0
	for _, path := range node.Dependencies {
		deps[i] = node.Graph[path]
		i++
	}

	return deps
}

func (node *DepNode) Dependents() []*DepNode {
	deps := []*DepNode{}

	for _, cnode := range node.Graph {
		if cnode.HasDependency(node.Path) {
			deps = append(deps, cnode)
		}
	}

	return deps
}

func (node *DepNode) DependentPaths() []string {
	deps := node.Dependents()
	depPaths := make([]string, len(deps))
	for i, dep := range deps {
		depPaths[i] = dep.Path
	}
	return depPaths
}

func (node *DepNode) HasDependency(path string) bool {
	for _, depPath := range node.Dependencies {
		if depPath == path {
			return true
		}
	}
	return false
}

func (node *DepNode) Detach(path string) error {
	if !node.HasDependency(path) {
		return fmt.Errorf(
			"dependency '%s' not found in %s's dependencies",
			path,
			node.Path,
		)
	}

	for k, v := range node.Dependencies {
		if v == path {
			delete(node.Dependencies, k)
			node.Graph.Delete(path)
			break
		}
	}

	return nil
}

func (graph DepGraph) Add(path string, deps map[string]string) {
	if node, ok := graph[path]; ok {
		for label, depPath := range deps {
			if !graph.Has(depPath) {
				graph.Add(depPath, map[string]string{})
			}

			node.Dependencies[label] = depPath
		}
	} else {
		graph[path] = &DepNode{
			Graph:        graph,
			Path:         path,
			Dependencies: map[string]string{},
		}

		graph.Add(path, deps)
	}
}

func (graph DepGraph) Has(path string) bool {
	_, hasDep := graph[path]
	return hasDep
}

func (graph DepGraph) Delete(path string) {
	if node, ok := graph[path]; !ok {
		return
	} else if len(node.Dependents()) > 0 {
		return
	}
	delete(graph, path)
}

func (graph DepGraph) Detach(path string, dep string) error {
	return graph[path].Detach(dep)
}
