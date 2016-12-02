package main

import (
	"fmt"
	"strings"

	"github.com/evanj/webconsole/deps"
)

// For flag docs see:
// https://github.com/google/closure-compiler/wiki/Using-NTI-(new-type-inference)
// need to
const closureVersion = "20161024"
const closureJAR = "closure-compiler-v" + closureVersion + ".jar"
const closureJARPath = buildOutputDir + "/" + closureJAR
const closureCompiler = "java -jar " + closureJARPath + " --emit_use_strict " +
	"--compilation_level ADVANCED --warning_level VERBOSE --new_type_inf " +
	"--jscomp_error '*' " +
	// must disable missing require: we don't use goog.require TODO: doesn't seem to work?
	"--jscomp_off missingRequire"
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

func addAll(set map[string]struct{}, values []string) {
	for _, v := range values {
		set[v] = struct{}{}
	}
}

// produces an output map where each js file's dependencies are flattened
func jsTransitiveDependencies(jsFiles map[string]*jsDependencies) map[string]*jsDependencies {
	// convert the imports to the format used by the deps package
	jsDeps := map[string][]string{}
	for file, dependencies := range jsFiles {
		jsDeps[file] = dependencies.imports
	}

	out := map[string]*jsDependencies{}
	for file := range jsFiles {
		transitiveDeps := deps.Transitive(jsDeps, file)

		// collect the set of externs corresponding to transitiveDeps
		externsSet := map[string]struct{}{}
		addAll(externsSet, jsFiles[file].externs)
		for _, dep := range transitiveDeps {
			addAll(externsSet, jsFiles[dep].externs)
		}

		externs := make([]string, 0, len(externsSet))
		for extern := range externsSet {
			externs = append(externs, extern)
		}

		out[file] = &jsDependencies{transitiveDeps, externs}
	}
	return out
}

func main() {
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

		"build/libapps/libdot/js/lib_f.js":             &jsDependencies{[]string{}, []string{"js/chrome_externs.js"}},
		"build/libapps/libdot/js/lib.js":               &jsDependencies{[]string{"build/libapps/libdot/js/lib_f.js"}, []string{}},
		"build/libapps/libdot/js/lib_utf8.js":          &jsDependencies{[]string{"build/libapps/libdot/js/lib_f.js"}, []string{}},
		"build/libapps/libdot/js/lib_wc.js":            &jsDependencies{[]string{"build/libapps/libdot/js/lib_f.js"}, []string{}},
		"build/libapps/libdot/js/lib_storage.js":       &jsDependencies{[]string{"build/libapps/libdot/js/lib_f.js"}, []string{}},
		"build/libapps/libdot/js/lib_storage_local.js": &jsDependencies{[]string{"build/libapps/libdot/js/lib_storage.js"}, []string{}},

		"js/htermwtf.js": &jsDependencies{[]string{
			"build/libapps/libdot/js/lib.js",
			// "build/libapps/libdot/js/lib_colors.js",
			// "build/libapps/libdot/js/lib_message_manager.js",
			// "build/libapps/libdot/js/lib_preference_manager.js",
			// "build/libapps/libdot/js/lib_resource.js",
			"build/libapps/libdot/js/lib_storage.js",
			// "build/libapps/libdot/js/lib_storage_chrome.js",
			"build/libapps/libdot/js/lib_storage_local.js",
			// "build/libapps/libdot/js/lib_storage_memory.js",
			// "build/libapps/libdot/js/lib_test_manager.js",
			"build/libapps/libdot/js/lib_utf8.js",
			"build/libapps/libdot/js/lib_wc.js",
		}, []string{}},
	}

	jsFlattenedDeps := jsTransitiveDependencies(jsFiles)
	jsInputs := []string{}
	jsCompiledTests := []string{}
	for inputPath, deps := range jsFlattenedDeps {
		// compile each file individually: ensures the dependencies are correct
		jsTarget := &jsModule{inputPath, deps.imports, deps.externs}
		targets = append(targets, jsTarget)

		// flatten the map to a lists for other targets
		jsInputs = append(jsInputs, inputPath)
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
