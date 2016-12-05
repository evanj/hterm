package deps

import (
	"reflect"
	"testing"
)

func TestOrderedStringSet(t *testing.T) {
	o := &orderedStringSet{}
	o.add("a")
	o.add("b")
	o.add("a")
	if !reflect.DeepEqual(o.values, []string{"a", "b"}) {
		t.Error(o.values)
	}

	o.moveOrPush("a")
	if !reflect.DeepEqual(o.values, []string{"b", "a"}) {
		t.Error(o.values)
	}
}

func TestTransitive(t *testing.T) {
	files := map[string][]string{
		"a.js": []string{},
		"b.js": []string{"a.js"},
		"c.js": []string{"b.js"},
		"d.js": []string{"a.js", "c.js"},
	}

	// order of imports matters!
	expected := []string{"a.js", "b.js", "c.js"}
	out := Transitive(files, "d.js")
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("%v != %v", expected, out)
	}
}

func TestTransitiveSort(t *testing.T) {
	g1 := map[string][]string{
		"a.js": []string{},
		"b.js": []string{"a.js"},
		"c.js": []string{"b.js"},
		"d.js": []string{"a.js", "c.js"},
	}

	tests := []struct {
		start    string
		expected []string
	}{
		{"a.js", []string{"a.js"}},
		{"b.js", []string{"a.js", "b.js"}},
		{"c.js", []string{"a.js", "b.js", "c.js"}},
		{"d.js", []string{"a.js", "b.js", "c.js", "d.js"}},
	}

	graph := NewGraph(g1)

	for i, test := range tests {
		out := graph.Dependencies(test.start)
		if !reflect.DeepEqual(out, test.expected) {
			t.Errorf("%d: start %s: got %v expected %v", i, test.start, out, test.expected)
		}
	}
}

func TestTransitiveComplex(t *testing.T) {
	g3 := map[string][]string{
		// 10 -> 21 22, 21 -> 31 32, 22 -> 38, 31 -> 41, 32 -> 41, 41 -> 22 50
		"10": []string{"21", "22"},
		"21": []string{"31", "32"},
		"22": []string{"38"},
		"31": []string{"41"},
		"32": []string{"41"},
		"41": []string{"22", "50"},
	}

	graph := NewGraph(g3)
	out := graph.Dependencies("10")
	// I think this is not unique, so I expect this will fail at some point
	expected := []string{"50", "38", "22", "41", "31", "32", "21", "10"}
	if !reflect.DeepEqual(out, expected) {
		t.Error(out, expected)
	}
}

func catchPanic(f func()) (paniced bool) {
	defer func() {
		r := recover()
		if r != nil {
			paniced = true
		}
	}()
	f()
	paniced = false
	return
}

func TestSortCycle(t *testing.T) {
	g := map[string][]string{
		"e.js": []string{"f.js"},
		"f.js": []string{"g.js"},
		"g.js": []string{"e.js"},

		"x.js": []string{"f.js"},
	}

	if !catchPanic(func() { topologicalSort(g) }) {
		t.Error("expected panic")
	}
}
