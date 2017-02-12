CLOSURE_COMPILER=java -jar build/closure-compiler-v20170124.jar --emit_use_strict --compilation_level ADVANCED --warning_level VERBOSE --new_type_inf  --jscomp_error accessControls --jscomp_error ambiguousFunctionDecl --jscomp_error checkEventfulObjectDisposal --jscomp_error checkRegExp --jscomp_error checkTypes --jscomp_error checkVars --jscomp_error commonJsModuleLoad --jscomp_error conformanceViolations --jscomp_error const --jscomp_error constantProperty --jscomp_error deprecated --jscomp_error deprecatedAnnotations --jscomp_error duplicateMessage --jscomp_error es3 --jscomp_error es5Strict --jscomp_error externsValidation --jscomp_error fileoverviewTags --jscomp_error functionParams --jscomp_error globalThis --jscomp_error internetExplorerChecks --jscomp_error invalidCasts --jscomp_error misplacedTypeAnnotation --jscomp_error missingGetCssName --jscomp_error missingOverride --jscomp_error missingPolyfill --jscomp_error missingProperties --jscomp_error missingProvide --jscomp_error missingReturn --jscomp_error msgDescriptions --jscomp_error newCheckTypes --jscomp_error nonStandardJsDocs --jscomp_error suspiciousCode --jscomp_error strictModuleDepCheck --jscomp_error typeInvalidation --jscomp_error undefinedNames --jscomp_error undefinedVars --jscomp_error unknownDefines --jscomp_error unusedLocalVariables --jscomp_error unusedPrivateMembers --jscomp_error uselessCode --jscomp_error useOfGoogBase --jscomp_error underscore --jscomp_error visibility

all: build/libapps build/closure-compiler-v20170124.jar build/js build/js/hterm_all.js build/../cmd/htermshell/static/htermshell.js build/../cmd/htermmenu/static/htermmenu.js build/__tests__/consolechannel-test.js build/js/consolechannel.js build/js/htermmenu.js build/js/htermshell.js build/uncompiled_tests.teststamp build/compiled_tests.teststamp

build/libapps:  | 
	git clone --depth 1 https://chromium.googlesource.com/apps/libapps build/libapps
	sed -i='' "s/date -u '+%a, %d %b %Y %T %z'/git log -1 --format=%cd/" build/libapps/hterm/concat/hterm_resources.concat

build/closure-compiler-v20170124.jar:  | 
	curl --location https://dl.google.com/closure-compiler/compiler-20170124.tar.gz | tar xvfz - -C build closure-compiler-v20170124.jar

build/js:  | 
	mkdir -p $@

build/js/hterm_all.js: build/closure-compiler-v20170124.jar build/libapps | build/js
	LIBDOT_SEARCH_PATH=$(pwd) build/libapps/libdot/bin/concat.sh -i build/libapps/hterm/concat/hterm_all.concat -o $@

build/../cmd/htermshell/static/htermshell.js: build/js/hterm_all.js build/js/htermshell.js | 
	cat $^ > $@

build/../cmd/htermmenu/static/htermmenu.js: build/js/hterm_all.js build/js/htermmenu.js | 
	cat $^ > $@

build/__tests__/consolechannel-test.js: js/consolechannel.js __tests__/consolechannel-test.js js/hterm_externs.js js/jasmine-2.0-externs.js js/node_externs.js build/closure-compiler-v20170124.jar | 
	$(CLOSURE_COMPILER) --js_output_file $@ --externs js/hterm_externs.js --externs js/jasmine-2.0-externs.js --externs js/node_externs.js js/consolechannel.js __tests__/consolechannel-test.js

build/js/consolechannel.js: js/consolechannel.js js/hterm_externs.js js/node_externs.js build/closure-compiler-v20170124.jar | 
	$(CLOSURE_COMPILER) --js_output_file $@ --externs js/hterm_externs.js --externs js/node_externs.js js/consolechannel.js

build/js/htermmenu.js: js/consolechannel.js js/htermmenu.js js/hterm_externs.js js/htermmenu_externs.js js/node_externs.js build/closure-compiler-v20170124.jar | 
	$(CLOSURE_COMPILER) --js_output_file $@ --externs js/hterm_externs.js --externs js/htermmenu_externs.js --externs js/node_externs.js js/consolechannel.js js/htermmenu.js

build/js/htermshell.js: js/consolechannel.js js/htermshell.js js/hterm_externs.js js/node_externs.js build/closure-compiler-v20170124.jar | 
	$(CLOSURE_COMPILER) --js_output_file $@ --externs js/hterm_externs.js --externs js/node_externs.js js/consolechannel.js js/htermshell.js

build/uncompiled_tests.teststamp: __tests__/consolechannel-test.js js/consolechannel.js js/htermmenu.js js/htermshell.js | 
	npm test
	touch $@

build/compiled_tests.teststamp: build/__tests__/consolechannel-test.js | 
	node_modules/.bin/jest '--config={"testRegex": "/build/__tests__/"}'
	touch $@

