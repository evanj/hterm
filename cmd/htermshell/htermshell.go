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

	flag.Parse()

	starter := hterm.NewSubprocessStarter(strings.Split(*cmd, " "))
	s := hterm.NewServer(starter)

	gopath := ""
	for _, env := range os.Environ() {
		fmt.Println("WTF", env)
		const gopathVar = "GOPATH="
		if strings.HasPrefix(env, gopathVar) {
			gopath = env[len(gopathVar):]
			break
		}
	}
	if gopath == "" {
		panic("Could not find GOPATH env var to locate static resources")
	}

	staticHandler := http.FileServer(http.Dir(gopath + "/" + gopathRelativeStaticDir))
	http.Handle("/", staticHandler)
	s.RegisterHandlers("/", http.DefaultServeMux)

	fmt.Printf("Listening on http://%s/\n", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		panic(err)
	}
}
