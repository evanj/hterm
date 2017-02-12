// we don't care about file modification timestamps but do want deterministic builds
//go:generate esc -pkg=$GOPACKAGE -o=static.go -prefix=static -modtime=1485035869 static
package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/evanj/hterm"
	"github.com/kr/pty"
)

const gopathRelativeStaticDir = "src/github.com/evanj/hterm/cmd/htermmenu/static"

var permittedCommands = []string{
	"ls", "vi", "man bash",
}

type indexTemplate struct {
	Commands []string
}

type executeTemplate struct {
	ConsoleExtra map[string]string
}

type server struct {
	staticHandler http.Handler
	index         *template.Template
	execute       *template.Template
}

func (s *server) rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("root", r.URL.Path)
	if r.URL.Path != "/" {
		s.staticHandler.ServeHTTP(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	values := &indexTemplate{permittedCommands}
	err := s.index.Execute(w, values)
	if err != nil {
		panic(err)
	}
}

func isPermittedCommand(command string) bool {
	for _, permitted := range permittedCommands {
		if command == permitted {
			return true
		}
	}
	return false
}

func (s *server) executeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("execute", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	command := r.PostFormValue("command")
	if !isPermittedCommand(command) {
		http.Error(w, "not permitted", http.StatusInternalServerError)
		return
	}

	values := &executeTemplate{map[string]string{"command": command}}
	err := s.execute.Execute(w, values)
	if err != nil {
		panic(err)
	}
}

// SessionStarter interface
func (s *server) Start(extraParams map[string]string) (*os.File, error) {
	// validate the command AGAIN: this is the real check
	command := extraParams["command"]
	if !isPermittedCommand(command) {
		return nil, errors.New("invalid command: " + command)
	}

	parts := strings.Split(command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	return pty.Start(cmd)
}

func readTemplate(fs http.FileSystem, name string) (*template.Template, error) {
	f, err := fs.Open(name)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(f)
	err2 := f.Close()
	if err != nil {
		return nil, err
	}
	if err2 != nil {
		return nil, err
	}
	return template.New(name).Parse(string(data))
}

func main() {
	addr := flag.String("addr", "localhost:8080", "Listening address e.g. :8080 for global")
	gopathStatic := flag.Bool("gopathStatic", false, "Open static resources from $GOPATH")

	flag.Parse()

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
		// compiled static resources
		fs = FS(false)
	}

	// load the templates
	index, err := readTemplate(fs, "index.html")
	if err != nil {
		panic(err)
	}
	execute, err := readTemplate(fs, "execute.html")
	if err != nil {
		panic(err)
	}
	s := &server{http.FileServer(fs), index, execute}
	htermServer := hterm.NewServer(s)

	http.HandleFunc("/", s.rootHandler)
	http.HandleFunc("/execute", s.executeHandler)
	htermServer.RegisterHandlers("/", http.DefaultServeMux)

	fmt.Printf("Listening on http://%s/\n", *addr)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		panic(err)
	}
}
