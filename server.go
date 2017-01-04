package hterm

import (
	"encoding/json"
	"errors"
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

var jsonEmptyObject = []byte("{}")

// SessionStarter creates a new session when Start is called.
type SessionStarter interface {
	// Start creates a new session with extraParams. The stream that is returned must produce
	// terminal output and consume it. It will be closed when the session is terminated or when
	// it times out.
	Start(extraParams map[string]string) (*os.File, error)
}

type subprocessStarter struct {
	command []string
}

// NewSubprocessStarter returns a SessionStarter that forks new subprocesses.
func NewSubprocessStarter(command []string) SessionStarter {
	return &subprocessStarter{command}
}

func (s *subprocessStarter) Start(extraParams map[string]string) (*os.File, error) {
	cmd := exec.Command(s.command[0], s.command[1:]...)
	return pty.Start(cmd)
}

type sessionState struct {
	id  string
	pty *os.File
}

type Server struct {
	mu       sync.Mutex
	sessions map[string]*sessionState
	starter  SessionStarter
}

func NewServer(starter SessionStarter) *Server {
	return &Server{sync.Mutex{}, map[string]*sessionState{}, starter}
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

func (s *Server) sessionWrapper(h customHandler) http.HandlerFunc {
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

				session.pty, err = s.starter.Start(req.Extra)
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

func (s *Server) writeHandler(w http.ResponseWriter, r *http.Request,
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
	w.Write(jsonEmptyObject)
	return nil
}

func (s *Server) setSizeHandler(w http.ResponseWriter, r *http.Request,
	session *sessionState, request *requestUnion) error {

	if request.Columns <= 0 || request.Rows <= 0 {
		return fmt.Errorf("invalid columns/rows: %d/%d", request.Columns, request.Rows)
	}

	log.Printf("setSize %d %d", request.Columns, request.Rows)
	err := setSize(session.pty, request.Columns, request.Rows)
	if err != nil {
		return err
	}
	w.Write(jsonEmptyObject)
	return nil
}

func (s *Server) readHandler(w http.ResponseWriter, r *http.Request,
	session *sessionState, request *requestUnion) error {
	// TODO: reuse buffer?
	buffer := make([]byte, 4096)
	n, err := session.pty.Read(buffer)
	if n > 0 {
		// handle any bytes read before any errors
		// TODO: this assumes calling Read again will return the error again
		log.Printf("readHandler: read %d bytes", n)

		// assume we can just convert this to UTF-8; TODO: how to handle escapes?
		resp := &readResponse{string(buffer[:n])}
		encoder := json.NewEncoder(w)
		return encoder.Encode(resp)
	}
	if err == io.EOF {
		log.Printf("readHandler: pty.Read returned EOF; session %s finished", session.id)
	}
	return err
}

func (s *Server) RegisterHandlers(path string, mux *http.ServeMux) {
	if len(path) == 0 || path[len(path)-1] != '/' {
		panic("path must end with /")
	}
	mux.HandleFunc(path+"write", s.sessionWrapper(s.writeHandler))
	mux.HandleFunc(path+"read", s.sessionWrapper(s.readHandler))
	mux.HandleFunc(path+"setSize", s.sessionWrapper(s.setSizeHandler))
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
