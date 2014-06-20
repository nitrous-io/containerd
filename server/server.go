package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nitrous-io/containerd/stdstream"

	"github.com/nitrous-io/containerd/agent"
	_ "github.com/nitrous-io/containerd/agent/docker_lxc"
)

const (
	defaultUploadMemory = 20 << 20 // 20 MB
	okResponse          = `{"status":"ok"}`
)

type Server struct {
	AgentType string // type of agent that we want to use i.e. lxc-attach
	Router    *mux.Router
	Shells    map[string]agent.Shell // list of current running shells
}

func (s *Server) getClient(r *http.Request) (agent.Client, error) {
	username := r.FormValue("username")
	return agent.GetClient(s.AgentType, mux.Vars(r)["id"], username)
}

func (s *Server) readFile(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// read the file
	filePath := r.FormValue("path")
	file, err := c.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) listDir(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dirPath := r.FormValue("path")
	fileInfos, err := c.ReadDir(dirPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(fileInfos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) writeFile(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.ParseMultipartForm(defaultUploadMemory); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filePath := r.FormValue("path")
	f, err := c.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer f.Close()
	for _, fhs := range r.MultipartForm.File {
		if len(fhs) <= 0 {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := copyUpload(f, fhs[0]); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Write([]byte(okResponse))
}

func (s *Server) renamePath(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oldPath := r.FormValue("from")
	newPath := r.FormValue("to")
	if err := c.Rename(oldPath, newPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(okResponse))
}

func (s *Server) deletePath(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	path := r.FormValue("path")
	if err := c.Remove(path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(okResponse))
}

func (s *Server) writeNewFile(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.ParseMultipartForm(defaultUploadMemory); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filePath := r.FormValue("path")
	f, err := c.CreateNew(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer f.Close()
	for _, fhs := range r.MultipartForm.File {
		if len(fhs) <= 0 {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := copyUpload(f, fhs[0]); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Write([]byte(okResponse))
}

func (s *Server) makeDir(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dirPath := r.FormValue("path")
	if err := c.Mkdir(dirPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(okResponse))
}

func (s *Server) newShell(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	shell, err := c.NewShell()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := randomString()
	s.Shells[id] = shell
	log.Println("new shell", id)
	w.Write([]byte(id))
}

func (s *Server) attachShell(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	term := r.FormValue("term")

	log.Println("attaching shell", id)

	width, err := strconv.ParseFloat(r.FormValue("width"), 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	height, err := strconv.ParseFloat(r.FormValue("height"), 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shell, ok := s.Shells[id]
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := shell.Open(term, width, height); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stdin, stdout, err := hijackServer(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if tcpc, ok := stdin.(*net.TCPConn); ok {
			tcpc.CloseWrite()
		} else {
			stdin.Close()
		}

		if tcpc, ok := stdout.(*net.TCPConn); ok {
			tcpc.CloseWrite()
		} else if closer, ok := stdout.(io.Closer); ok {
			closer.Close()
		}
	}()

	// flush the output
	fmt.Fprintf(stdout, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.gomux.raw-stream\r\n\r\n")
	err = <-shell.Attach(stdin, stdout)
	log.Println("error", err)
}

func (s *Server) resizeShell(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	width, err := strconv.ParseFloat(r.FormValue("width"), 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	height, err := strconv.ParseFloat(r.FormValue("height"), 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("resize", id, width, height)

	shell, ok := s.Shells[id]
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := shell.Resize(width, height); err != nil {
		log.Println("resize err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(okResponse))
}

func (s *Server) runCommand(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient(r)
	if err != nil {
		log.Println("client", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cmd := c.NewCommand(r.FormValue("cmd"))
	_, outR, errR, err := cmd.AllPipe()
	if err != nil {
		log.Println("pipe", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		stdstream.NewStdStream(w, map[string]io.Reader{
			"stdout": outR,
			"stderr": errR,
		}).Flush()
	}()

	if err := cmd.Start(); err != nil {
		log.Println("start", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Println("wait", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreateRouter() *mux.Router {
	r := mux.NewRouter()

	// setup http handler
	m := map[string]map[string]http.HandlerFunc{
		"GET": {
			"/containers/{id:.*}/file": s.readFile,
			"/containers/{id:.*}/dir":  s.listDir,
		},
		"PUT": {
			"/containers/{id:.*}/file":   s.writeFile,
			"/containers/{id:.*}/rename": s.renamePath,
		},
		"POST": {
			"/containers/{id:.*}/file":    s.writeNewFile,
			"/containers/{id:.*}/dir":     s.makeDir,
			"/containers/{id:.*}/command": s.runCommand,
			"/containers/{id:.*}/shell":   s.newShell,
			"/shell/{id:.*}/attach":       s.attachShell,
			"/shell/{id:.*}/resize":       s.resizeShell,
		},
		"DELETE": {
			"/containers/{id:.*}/path": s.deletePath,
		},
	}
	for method, routes := range m {
		for route, handler := range routes {
			localRoute := route
			localHandler := handler
			localMethod := method
			r.HandleFunc(localRoute, localHandler).Methods(localMethod)
		}
	}

	return r
}

func NewServer(agentType string) *Server {
	s := &Server{AgentType: agentType}
	s.Shells = map[string]agent.Shell{}
	s.Router = s.CreateRouter()
	return s
}

func copyUpload(dest io.WriteCloser, uploaded *multipart.FileHeader) error {
	defer dest.Close()
	from, err := uploaded.Open()
	if err != nil {
		return err
	}
	_, err = io.Copy(dest, from)
	return err
}

func hijackServer(w http.ResponseWriter) (io.ReadCloser, io.Writer, error) {
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return nil, nil, err
	}
	// Flush the options to make sure the client sets the raw mode
	conn.Write([]byte{})
	return conn, conn, nil
}
