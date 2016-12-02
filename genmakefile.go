package main

import (
	"reflect"
	"strings"

	"fmt"
)

// For flag docs see:
// https://github.com/google/closure-compiler/wiki/Using-NTI-(new-type-inference)
const closureVersion = "20161024"
const closureJAR = "closure-compiler-v" + closureVersion + ".jar"
const closureJARPath = buildOutputDir + "/" + closureJAR
const closureCompiler = "java -jar " + closureJARPath + " --emit_use_strict --compilation_level ADVANCED --warning_level VERBOSE --new_type_inf --jscomp_error '*'"
const jest = "node_modules/.bin/jest"
const buildOutputDir = "build"
const jsTestPrefix = "__tests__/"

func buildOutput(name string) string {
	return buildOutputDir + "/" + name
}

type target interface {
	output() string
	inputs() []string
	commands() []string
}

type jsModule struct {
	path         string
	dependencies []string
	externs      []string
}

func (j *jsModule) output() string {
	return buildOutput(j.path)
}

func (j *jsModule) compiledInputs() []string {
	var out []string = nil
	out = append(out, j.dependencies...)
	// the "original" file must go after all dependencies so the definitions are available
	out = append(out, j.path)
	return out
}

func (j *jsModule) inputs() []string {
	out := j.compiledInputs()
	out = append(out, j.externs...)
	return append(out, closureJARPath)
}

func (j *jsModule) commands() []string {
	command := "$(CLOSURE_COMPILER) --js_output_file $@ "
	for _, extern := range j.externs {
		command += "--externs " + extern + " "
	}
	command += strings.Join(j.compiledInputs(), " ")
	return []string{command}
}

type staticTarget struct {
	out  string
	ins  []string
	cmds []string
}

func (s *staticTarget) output() string {
	return buildOutput(s.out)
}

func (s *staticTarget) inputs() []string {
	return s.ins
}

func (s *staticTarget) commands() []string {
	return s.cmds
}

type jsDependencies struct {
	imports []string
	externs []string
}

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

func reverse(values []string) []string {
	for i := 0; i < len(values)/2; i++ {
		j := len(values) - i - 1
		values[i], values[j] = values[j], values[i]
	}
	return values
}

func transitiveDependencies(files map[string]*jsDependencies, input string) *jsDependencies {
	// imports must be ordered from leaves up to the root
	imports := &orderedStringSet{[]string{}}
	externs := &orderedStringSet{[]string{}}

	// TODO: prevent self-import?
	toVisit := []string{input}
	for len(toVisit) > 0 {
		// pop the next file to visit
		in := toVisit[len(toVisit)-1]
		toVisit = toVisit[:len(toVisit)-1]

		// collect all of its imports and externs
		deps := files[in]
		for _, i := range deps.imports {
			if !imports.contains(i) {
				// we have not seen this import yet: must recursively visit
				imports.add(i)
				toVisit = append(toVisit, i)
			}
		}
		externs.addAll(deps.externs)
	}

	reverse(imports.values)
	return &jsDependencies{imports.values, externs.values}
}

// TODO: make this a real test
func testDependencies() {
	o := &orderedStringSet{}
	o.add("a")
	o.add("b")
	o.add("a")
	if !reflect.DeepEqual(o.values, []string{"a", "b"}) {
		panic(fmt.Sprintf("%v", o.values))
	}

	files := map[string]*jsDependencies{
		"base1.js": &jsDependencies{[]string{}, []string{"e1.js"}},
		"base2.js": &jsDependencies{[]string{}, []string{"e1.js", "e2.js"}},
		"a.js":     &jsDependencies{[]string{"base1.js"}, []string{"e3.js"}},
		"b.js":     &jsDependencies{[]string{"a.js", "base2.js"}, []string{}},
	}

	expected := files["base1.js"]
	deps := transitiveDependencies(files, "base1.js")
	if !reflect.DeepEqual(deps, expected) {
		panic(fmt.Sprintf("%v ;;; %v != %v", reflect.DeepEqual(deps.imports, expected.imports), deps, files["base1.js"]))
	}

	// order of imports matters!
	expectedImports := []string{"base1.js", "base2.js", "a.js"}
	expectedExterns := []string{"e1.js", "e2.js", "e3.js"}
	deps = transitiveDependencies(files, "b.js")
	if !reflect.DeepEqual(expectedImports, deps.imports) {
		panic(fmt.Sprintf("%v != %v", expectedImports, deps.imports))
	}
	if !reflect.DeepEqual(expectedExterns, deps.externs) {
		panic(fmt.Sprintf("%v != %v", expectedExterns, deps.externs))
	}
}

func main() {
	testDependencies()

	targets := []target{
		&staticTarget{"libapps", []string{},
			[]string{"git clone --depth 1 https://chromium.googlesource.com/apps/libapps build/libapps"}},
		&staticTarget{closureJAR, []string{},
			[]string{"curl --location https://dl.google.com/closure-compiler/compiler-" + closureVersion + ".tar.gz | tar xvf - -C " + buildOutputDir + " *.jar"}},
		&staticTarget{"hterm_all.js", []string{closureJARPath, buildOutput("libapps")},
			[]string{"LIBDOT_SEARCH_PATH=$(pwd) build/libapps/libdot/bin/concat.sh -i build/libapps/hterm/concat/hterm_all.concat -o build/hterm_all.js"}},
	}

	jsFiles := map[string]*jsDependencies{
		"js/consolechannel.js":             &jsDependencies{[]string{}, []string{"js/hterm_externs.js", "js/node_externs.js"}},
		"__tests__/consolechannel-test.js": &jsDependencies{[]string{"js/consolechannel.js"}, []string{"js/jasmine-2.0-externs.js"}},
	}

	jsInputs := []string{}
	jsCompiledTests := []string{}
	for inputPath := range jsFiles {
		// compile each file individually: ensures the dependencies are correct
		deps := transitiveDependencies(jsFiles, inputPath)
		jsTarget := &jsModule{inputPath, deps.imports, deps.externs}
		targets = append(targets, jsTarget)
		jsInputs = append(jsInputs, inputPath)

		// compile the tests
		if strings.HasPrefix(inputPath, jsTestPrefix) {
			jsCompiledTests = append(jsCompiledTests, jsTarget.output())
		}
	}

	// run all tests uncompiled: assume they depend on all .js files
	// TODO: this should be the transitive dependencies of the tests themselves, but whatever
	targets = append(targets, &staticTarget{"uncompiled_tests.teststamp", jsInputs,
		[]string{"npm test", "touch $@"}})

	// run all the compiled tests
	targets = append(targets, &staticTarget{"compiled_tests.teststamp", jsCompiledTests,
		[]string{jest + ` '--config={"testRegex": "/build/__tests__/"}'`, "touch $@"}})

	allTargets := ""
	targetOutput := ""
	for i, target := range targets {
		if i != 0 {
			allTargets += " "
		}
		allTargets += target.output()

		targetOutput += target.output() + ": " + strings.Join(target.inputs(), " ") + "\n"
		for _, cmd := range target.commands() {
			targetOutput += "\t" + cmd + "\n"
		}
		targetOutput += "\n"
	}

	fmt.Printf("CLOSURE_COMPILER=%s\n", closureCompiler)
	fmt.Printf("\nall: %s\n\n", allTargets)
	fmt.Print(targetOutput)
}
