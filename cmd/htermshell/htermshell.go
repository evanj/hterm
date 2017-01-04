package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/evanj/hterm"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "Listening address e.g. :8080 for global")
	cmd := flag.String("cmd", "bash -l", "Command to run (no shell variable expansion)")
	flag.Parse()

	starter := hterm.NewSubprocessStarter(strings.Split(*cmd, " "))
	s := hterm.NewServer(starter)

	staticHandler := http.FileServer(http.Dir("build/js"))
	http.Handle("/", staticHandler)
	s.RegisterHandlers("/", http.DefaultServeMux)

	fmt.Printf("Listening on http://%s/\n", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		panic(err)
	}
}
