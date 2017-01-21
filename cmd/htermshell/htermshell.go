// we don't care about file modification timestamps but do want deterministic builds
//go:generate esc -pkg=$GOPACKAGE -o=static.go -prefix=static -modtime=1485035869 static
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/evanj/hterm"
)

const gopathRelativeStaticDir = "src/github.com/evanj/hterm/cmd/htermshell/static"

func main() {
	addr := flag.String("addr", "localhost:8080", "Listening address e.g. :8080 for global")
	cmd := flag.String("cmd", "bash -l", "Command to run (no shell variable expansion)")
	gopathStatic := flag.Bool("gopathStatic", false, "Open static resources from $GOPATH")

	flag.Parse()

	starter := hterm.NewSubprocessStarter(strings.Split(*cmd, " "))
	s := hterm.NewServer(starter)

	// Use the "real" http.FileSystem since we don't want to depend on the current working directory
	var fs http.FileSystem
	if *gopathStatic {
		gopath := ""
		for _, env := range os.Environ() {
			const gopathVar = "GOPATH="
			if strings.HasPrefix(env, gopathVar) {
				gopath = env[len(gopathVar):]
				break
			}
		}
		if gopath == "" {
			panic("Could not find GOPATH env var to locate static resources")
		}
		fs = http.Dir(gopath + "/" + gopathRelativeStaticDir)
	} else {
		fs = FS(false)
	}

	staticHandler := http.FileServer(fs)
	http.Handle("/", staticHandler)
	s.RegisterHandlers("/", http.DefaultServeMux)

	fmt.Printf("Listening on http://%s/\n", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		panic(err)
	}
}
