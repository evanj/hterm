package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/kr/pty"
)

type sessionState struct {
	id  string
	pty *os.File
}

type server struct {
	sessions     map[string]*sessionState
	startCommand []string
}

type requestBase struct {
	SessionId string `json:"session_id"`
	Query     string `json:"query"`
}

type writeRequest struct {
	requestBase
	Data string `json:"data"`
}

type setSizeRequest struct {
	requestBase
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

type readResponse struct {
	Data string `json:"data"`
}

type customHandler func(w http.ResponseWriter, r *http.Request,
	session *sessionState, data []byte) error

// TODO: Requires synchronization
func (s *server) sessionWrapper(h customHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := func() error {
			if r.Method != http.MethodPost {
				return fmt.Errorf("invalid method %s", r.Method)
			}
			data, err := ioutil.ReadAll(r.Body)
			log.Println("WTF???", string(data))
			err2 := r.Body.Close()
			if err != nil {
				return err
			}
			if err2 != nil {
				return err2
			}
			// json decode just the session
			req := &requestBase{}
			err = json.Unmarshal(data, req)
			if err != nil {
				return err
			}

			if req.SessionId == "" {
				return errors.New("required field session_id is missing")
			}

			session := s.sessions[req.SessionId]
			if session == nil {
				log.Printf("starting new session %s", req.SessionId)
				session = &sessionState{id: req.SessionId}

				startCommand := s.startCommand
				if req.Query != "" {
					values, err := url.ParseQuery(req.Query)
					if err != nil {
						return err
					}
					if values.Get("instance") == "" {
						return errors.New("instance parameter is required")
					}
					instance := values.Get("instance")
					user := "root"
					if values.Get("user") != "" {
						user = values.Get("user")
					}

					startCommand = []string{
						"mysql", "--socket=/cloudsql/" + instance, "--user=" + user, "--password"}
				}

				cmd := exec.Command(startCommand[0], startCommand[1:]...)
				session.pty, err = pty.Start(cmd)
				if err != nil {
					return err
				}
				s.sessions[req.SessionId] = session
			}

			// pass on the request
			log.Printf("%s session %s", r.URL.Path, session.id)
			return h(w, r, session, data)
		}()
		if err != nil {
			log.Printf("Error: %s: %s", r.URL.Path, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// func rootHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("rootHandler %s", r.URL.Path)
// 	if r.URL.Path != "/" {
// 		http.NotFound(w, r)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
// 	w.Write([]byte("hello root"))
// }

func (s *server) writeHandler(w http.ResponseWriter, r *http.Request,
	session *sessionState, data []byte) error {
	request := &writeRequest{}
	err := json.Unmarshal(data, request)
	if err != nil {
		return err
	}

	if request.Data == "" {
		return errors.New("missing required field data")
	}

	log.Printf("received %s\n", request.Data)
	// TODO: figure out how to transform this to the appropriate format
	n, err := session.pty.Write([]byte(request.Data))
	if err != nil {
		return err
	}
	log.Printf("writeHandler: wrote %d bytes", n)
	return nil
}

func (s *server) setSizeHandler(w http.ResponseWriter, r *http.Request,
	session *sessionState, data []byte) error {
	request := &setSizeRequest{}
	err := json.Unmarshal(data, request)
	if err != nil {
		return err
	}

	if request.Columns <= 0 || request.Rows <= 0 {
		return fmt.Errorf("invalid columns/rows: %d/%d", request.Columns, request.Rows)
	}

	log.Printf("setSize %d %d", request.Columns, request.Rows)
	err = setSize(session.pty, request.Columns, request.Rows)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) readHandler(w http.ResponseWriter, r *http.Request,
	session *sessionState, data []byte) error {
	// TODO: reuse buffer; loop appropriately
	buffer := make([]byte, 4096)
	n, err := session.pty.Read(buffer)
	if err != nil {
		return err
	}
	if n == len(buffer) {
		panic("filled the buffer TODO: implement")
	}
	if n <= 0 {
		panic("WTF n <= 0")
	}
	log.Printf("readHandler: read %d bytes", n)

	// assume we can just convert this to UTF-8; TODO: how to handle escapes?
	resp := &readResponse{string(buffer[:n])}
	encoder := json.NewEncoder(w)
	err = encoder.Encode(resp)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// start a subprocess
	s := &server{sessions: make(map[string]*sessionState)}
	// s.startCommand = []string("mysql", "--user=root", "--socket=/tmp/triggeredmail:us-central1:statpoller", "costs")
	s.startCommand = []string{"bash"}

	staticHandler := http.FileServer(http.Dir("static"))
	http.Handle("/", staticHandler)
	http.HandleFunc("/write", s.sessionWrapper(s.writeHandler))
	http.HandleFunc("/read", s.sessionWrapper(s.readHandler))
	http.HandleFunc("/setSize", s.sessionWrapper(s.setSizeHandler))

	const listenHostPort = ":8080"
	fmt.Printf("Listening on http://%s/\n", listenHostPort)
	err := http.ListenAndServe(listenHostPort, nil)
	if err != nil {
		panic(err)
	}
}

// Setsize resizes pty t to s.
// From https://github.com/kr/pty/pull/39/files
func setSize(t *os.File, columns int, rows int) error {
	const pixelsPerColumn = 640 / 80
	const pixelsPerRow = 480 / 24
	ws := &winsize{uint16(rows), uint16(columns),
		uint16(pixelsPerRow * columns), uint16(pixelsPerRow * rows)}
	return windowRectCall(ws, t.Fd(), syscall.TIOCSWINSZ)
}

// Winsize describes the terminal size.
type winsize struct {
	Rows uint16 // ws_row: Number of rows (in cells)
	Cols uint16 // ws_col: Number of columns (in cells)
	X    uint16 // ws_xpixel: Width in pixels
	Y    uint16 // ws_ypixel: Height in pixels
}

func windowRectCall(ws *winsize, fd, a2 uintptr) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		a2,
		uintptr(unsafe.Pointer(ws)),
	)
	if errno != 0 {
		return syscall.Errno(errno)
	}
	return nil
}
