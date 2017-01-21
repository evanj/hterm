package main

import (
	"fmt"
	"sort"
)

type orderedStringSet struct {
	values []string
}

func (s *orderedStringSet) contains(value string) bool {
	for _, v := range s.values {
		if v == value {
			return true
		}
	}
	return false
}

func (s *orderedStringSet) add(value string) {
	if !s.contains(value) {
		s.values = append(s.values, value)
	}
}

func (s *orderedStringSet) addAll(values []string) {
	for _, v := range values {
		s.add(v)
	}
}

func (s *orderedStringSet) isEmpty() bool {
	return len(s.values) == 0
}

func (s *orderedStringSet) pop() string {
	last := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return last
}

func (s *orderedStringSet) checkPush(value string) {
	if s.contains(value) {
		panic("set already contains value: " + value)
	}
	s.values = append(s.values, value)
}

func (s *orderedStringSet) moveOrPush(value string) {
	for i, v := range s.values {
		if v == value {
			// move this element to the end
			s.values[len(s.values)-1], s.values[i] = s.values[i], s.values[len(s.values)-1]
			return
		}
	}

	s.values = append(s.values, value)
}

func Transitive(files map[string][]string, start string) []string {
	// order matters. eg. b requires a; c requires b; d requires (b, c) OUTPUT: d, c, b, a
	// we must visit in breadth first order from root to the leaves, so intermediate state could be: d, b, a
	// when we visit c, we must find the minimum index of its dependencies and insert it BEFORE

	// ensure that all of a file's dependencies are added before adding the file
	// to do that: push unvisited dependencies on a stack, then

	// TODO: prevent self-import? detect cycles?
	visitStack := &orderedStringSet{[]string{start}}
	outputs := &orderedStringSet{}

	for !visitStack.isEmpty() {
		next := visitStack.pop()
		deps := files[next]

		unvisitedDependencies := []string{}
		for _, dep := range deps {
			if !outputs.contains(dep) {
				unvisitedDependencies = append(unvisitedDependencies, dep)
			}
		}

		if len(unvisitedDependencies) == 0 {
			// we've resolved all the dependencies: put it in the output
			outputs.add(next)
		} else {
			// we need to revisit this AFTER we visit the unvisited dependencies
			// TODO: detect cycles to avoid an infinite loop
			visitStack.checkPush(next)
			for _, dep := range unvisitedDependencies {
				visitStack.moveOrPush(dep)
			}
		}
	}

	return outputs.values[:len(outputs.values)-1]
}

func pop(set map[string]struct{}) string {
	for value, _ := range set {
		delete(set, value)
		return value
	}
	panic("cannot pop from empty set")
}

func index(values []string, value string) int {
	for i, v := range values {
		if v == value {
			return i
		}
	}
	return -1
}

type unorderedSet struct {
	values map[string]struct{}
}

func newUnorderedSet() *unorderedSet {
	return &unorderedSet{map[string]struct{}{}}
}

func (set *unorderedSet) add(value string) {
	set.values[value] = struct{}{}
}

func (set *unorderedSet) remove(value string) bool {
	_, exists := set.values[value]
	if exists {
		delete(set.values, value)
	}
	return exists
}

func (set *unorderedSet) pop() string {
	for value, _ := range set.values {
		delete(set.values, value)
		return value
	}
	panic("cannot pop from empty set")
}

func (set *unorderedSet) isEmpty() bool {
	return len(set.values) == 0
}

func topologicalSort(outgoingEdges map[string][]string) []string {
	// need the incoming edges: copy the graph
	incomingEdges := map[string]*unorderedSet{}
	for src, dests := range outgoingEdges {
		set := incomingEdges[src]
		if set == nil {
			set = newUnorderedSet()
			incomingEdges[src] = set
		}
		for _, dest := range dests {
			set := incomingEdges[dest]
			if set == nil {
				set = newUnorderedSet()
				incomingEdges[dest] = set
			}
			set.add(src)
		}
	}

	noIncoming := &unorderedSet{map[string]struct{}{}}
	for dest, srcs := range incomingEdges {
		if srcs.isEmpty() {
			noIncoming.add(dest)
		}
	}

	output := []string{}
	for !noIncoming.isEmpty() {
		// everything that currently has no incoming edges has an equal weight:
		// collect them all and sort by key
		keys := []string{}
		for node := range noIncoming.values {
			keys = append(keys, node)
		}
		sort.Strings(keys)

		for _, node := range keys {
			noIncoming.remove(node)
			output = append(output, node)

			// remove the edge from all destinations
			for _, edge := range outgoingEdges[node] {
				success := incomingEdges[edge].remove(node)
				if !success {
					panic("mismatch between incoming and outgoing edges")
				}
				if incomingEdges[edge].isEmpty() {
					noIncoming.add(edge)
				}
			}
		}
	}

	if len(output) != len(incomingEdges) {
		panic(fmt.Sprintf("cycle detected: edges still remaining %v %v", incomingEdges, output))
	}
	return output
}

type Graph struct {
	topologicalIndex map[string]int
	dependencies     map[string][]string
}

func NewGraph(dependencies map[string][]string) *Graph {
	g := &Graph{map[string]int{}, dependencies}

	topological := topologicalSort(dependencies)
	for i, node := range topological {
		g.topologicalIndex[node] = i
	}
	return g
}

type topologicalSorter struct {
	nodes   []string
	indexes map[string]int
}

func (t *topologicalSorter) Len() int {
	return len(t.nodes)
}

func (t *topologicalSorter) Swap(i, j int) {
	t.nodes[i], t.nodes[j] = t.nodes[j], t.nodes[i]
}

func (t *topologicalSorter) Less(i, j int) bool {
	return t.indexes[t.nodes[i]] > t.indexes[t.nodes[j]]
}

func (g *Graph) Dependencies(start string) []string {
	// collect all the dependencies
	fileDeps := orderedStringSet{[]string{start}}
	for i := 0; i < len(fileDeps.values); i++ {
		node := fileDeps.values[i]
		for _, dep := range g.dependencies[node] {
			if !fileDeps.contains(dep) {
				fileDeps.add(dep)
			}
		}
	}

	sort.Sort(&topologicalSorter{fileDeps.values, g.topologicalIndex})
	return fileDeps.values
}
