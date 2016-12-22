package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"github.com/kr/pty"
)

// SessionFactory creates a new session with extraParams. It returns an io.ReadWriteCloser that
// produces and consumes terminal input and output, or an error.
type SessionFactory func(extraParams map[string]string) (io.ReadWriteCloser, error)

type sessionState struct {
	id  string
	pty *os.File
}

type server struct {
	mu           sync.Mutex
	sessions     map[string]*sessionState
	startCommand []string
}

// Union for write, read, and setSize requests
type requestUnion struct {
	// common parameters
	SessionId string            `json:"session_id"`
	Extra     map[string]string `json:"extra"`

	// write
	Data string `json:"data"`

	// setSize
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

type readResponse struct {
	Data string `json:"data"`
}

type customHandler func(w http.ResponseWriter, r *http.Request,
	session *sessionState, req *requestUnion) error

func (s *server) sessionWrapper(h customHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := func() error {
			if r.Method != http.MethodPost {
				return fmt.Errorf("invalid method %s", r.Method)
			}
			data, err := ioutil.ReadAll(r.Body)
			log.Printf("sessionWrapper read message: %s", string(data))
			err2 := r.Body.Close()
			if err != nil {
				return err
			}
			if err2 != nil {
				return err2
			}

			// json decode the request
			req := &requestUnion{}
			err = json.Unmarshal(data, req)
			if err != nil {
				return err
			}

			if req.SessionId == "" {
				return errors.New("required field session_id is missing")
			}

			s.mu.Lock()
			session := s.sessions[req.SessionId]
			s.mu.Unlock()

			if session == nil {
				log.Printf("creating new session id %s", req.SessionId)
				session = &sessionState{id: req.SessionId}

				startCommand := s.startCommand
				// if req. != "" {
				// 	values, err := url.ParseQuery(req.Query)
				// 	if err != nil {
				// 		return err
				// 	}
				// 	if values.Get("instance") == "" {
				// 		return errors.New("instance parameter is required")
				// 	}
				// 	instance := values.Get("instance")
				// 	user := "root"
				// 	if values.Get("user") != "" {
				// 		user = values.Get("user")
				// 	}

				// 	startCommand = []string{
				// 		"mysql", "--socket=/cloudsql/" + instance, "--user=" + user, "--password"}
				// }

				cmd := exec.Command(startCommand[0], startCommand[1:]...)
				session.pty, err = pty.Start(cmd)
				if err != nil {
					return err
				}

				s.mu.Lock()
				s.sessions[req.SessionId] = session
				s.mu.Unlock()
			}

			// pass on the request to the real handler
			log.Printf("%s session %s", r.URL.Path, session.id)
			return h(w, r, session, req)
		}()
		if err != nil {
			log.Printf("Error: %s: %s", r.URL.Path, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *server) writeHandler(w http.ResponseWriter, r *http.Request,
	session *sessionState, request *requestUnion) error {

	if request.Data == "" {
		return errors.New("write request missing required data")
	}

	log.Printf("writeHandler data: %s\n", request.Data)
	n, err := session.pty.Write([]byte(request.Data))
	if err != nil {
		return err
	}
	log.Printf("writeHandler: wrote %d bytes", n)
	return nil
}

func (s *server) setSizeHandler(w http.ResponseWriter, r *http.Request,
	session *sessionState, request *requestUnion) error {

	if request.Columns <= 0 || request.Rows <= 0 {
		return fmt.Errorf("invalid columns/rows: %d/%d", request.Columns, request.Rows)
	}

	log.Printf("setSize %d %d", request.Columns, request.Rows)
	return setSize(session.pty, request.Columns, request.Rows)
}

func (s *server) readHandler(w http.ResponseWriter, r *http.Request,
	session *sessionState, request *requestUnion) error {
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

func startSubprocess(command []string) (io.ReadWriteCloser, error) {
	cmd := exec.Command(command[0], command[1:]...)
	return pty.Start(cmd)
}

func main() {
	addr := flag.String("addr", "localhost:8080", "Socket address to listen on e.g. :8080 for global")
	s := &server{sync.Mutex{}, make(map[string]*sessionState), []string{"bash"}}

	staticHandler := http.FileServer(http.Dir("build/js"))
	http.Handle("/", staticHandler)
	http.HandleFunc("/write", s.sessionWrapper(s.writeHandler))
	http.HandleFunc("/read", s.sessionWrapper(s.readHandler))
	http.HandleFunc("/setSize", s.sessionWrapper(s.setSizeHandler))

	fmt.Printf("Listening on http://%s/\n", *addr)
	err := http.ListenAndServe(*addr, nil)
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
