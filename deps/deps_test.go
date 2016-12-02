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
