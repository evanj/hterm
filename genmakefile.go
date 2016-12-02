package main

import (
	"strings"

	"fmt"
)

// For flag docs see:
// https://github.com/google/closure-compiler/wiki/Using-NTI-(new-type-inference)
const closureVersion = "20161024"
const closureJAR = "closure-compiler-v" + closureVersion + ".jar"
const closureJARPath = buildOutputDir + "/" + closureJAR
const closureCompiler = "java -jar " + closureJAR + " --emit_use_strict --compilation_level ADVANCED --warning_level VERBOSE --new_type_inf --jscomp_error '*'"
const jest = "node_modules/.bin/jest"
const buildOutputDir = "build"
const jsTestPrefix = "__tests__/"

// all: build/hterm_all.js

// build/hterm_all.js: build/libapps
//   LIBDOT_SEARCH_PATH=$(pwd) build/libapps/libdot/bin/concat.sh -i build/libapps/hterm/concat/hterm_all.concat -o build/hterm_all.js

// build/libapps:
//   git clone --depth 1 https://chromium.googlesource.com/apps/libapps build/libapps

// build/closure-compiler-v20161024.jar:
//   curl --location https://dl.google.com/closure-compiler/compiler-20161024.tar.gz | tar xvf - *.jar

// build/hterm_compiled.js: build/closure-compiler-v20161024.jar build/hterm_all.js
//   java -jar build/closure-compiler-v20161024.jar --emit_use_strict --compilation_level ADVANCED --warning_level VERBOSE --new_type_inf --jscomp_error '*' --js build/hterm_all.js > build/hterm_compiled.js

// build/__tests__/consolechannel-test.js: js/consolechannel.js __tests__/consolechannel-test.js
//   $(CLOSURE_COMPILER) $^ --js_output_file $@ --externs js/node_externs.js --externs js/jasmine-2.0-externs.js --externs js/hterm_externs.js

// build/__tests__/consolechannel-test.js.teststamp: build/__tests__/consolechannel-test.js
//   $(JEST) $<
//   touch $@

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
	var out []string = []string{j.path}
	return append(out, j.dependencies...)
}

func (j *jsModule) inputs() []string {
	out := j.compiledInputs()
	return append(out, j.externs...)
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

// js tests are special: they require multiple targets
// func makeJSTest(jsFiles map[string]*jsDependencies, input string, dependencies []string) []target {
// 	out := []target{}

// 	// rule to run the test, without compilation
// 	allJSDependencies := []string{input}
// 	allJSDependencies = append(allJSDependencies, dependencies...)
// 	out = append(out, &staticTarget{input + ".teststamp", allJSDependencies,
// 		[]string{"$(JEST) " + input, "touch $@"}})

// 	// rule to compile the test
// 	out = append(out, &jsModule{input, dependencies})
// 	allJSDependencies := []string{input}
// 	allJSDependencies = append(allJSDependencies, dependencies...)
// 	out = append(out, &staticTarget{input + ".teststamp", allJSDependencies,
// 		[]string{"$(JEST) " + input, "touch $@"}})

// 	return out
// }

func main() {
	fmt.Printf("CLOSURE_COMPILER=%s\n", closureCompiler)

	targets := []target{
		&staticTarget{"libapps", []string{},
			[]string{"git clone --depth 1 https://chromium.googlesource.com/apps/libapps build/libapps"}},
		&staticTarget{closureJAR, []string{},
			[]string{"curl --location https://dl.google.com/closure-compiler/compiler-" + closureVersion + ".tar.gz | tar xvf - *.jar"}},
		&staticTarget{"hterm_all.js", []string{closureJARPath, buildOutput("libapps")},
			[]string{"LIBDOT_SEARCH_PATH=$(pwd) build/libapps/libdot/bin/concat.sh -i build/libapps/hterm/concat/hterm_all.concat -o build/hterm_all.js"}},
	}

	jsFiles := map[string]*jsDependencies{
		"js/consolechannel.js":             &jsDependencies{[]string{}, []string{"js/hterm_externs.js", "js/node_externs.js"}},
		"__tests__/consolechannel-test.js": &jsDependencies{[]string{"js/consolechannel.js"}, []string{"js/jasmine-2.0-externs.js"}},
	}

	jsInputs := []string{}
	jsCompiledTests := []string{}
	for inputPath, deps := range jsFiles {
		// compile each file individually: ensures the dependencies are correct
		// TODO: compute the *transitive dependencies* for the rule; this is just the direct dependencies
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

	// targets := []target{}
	// for _, t := range jsTargets {
	// 	targets = append(targets, t)
	// }
	// for _, t := range staticTargets {
	// 	targets = append(targets, t)
	// }

	// targets = append(targets, makeJSTest("__tests__/consolechannel-test.js", []string{"js/consolechannel.js"})...)

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

	fmt.Printf("\nall: %s\n\n", allTargets)

	fmt.Print(targetOutput)
}
