package deps

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
