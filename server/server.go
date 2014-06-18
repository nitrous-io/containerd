package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nitrous-io/agentd/agent"
)

type Server struct {
	AgentType string // type of agent that we want to use i.e. lxc-attach
	Router    *mux.Router
	Clients   map[string]agent.Client // list of current running clients
}

func (s *Server) readFile(w http.ResponseWriter, r *http.Request) {
	c, err := agent.GetClient(s.AgentType)
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
}

func (s *Server) writeFile(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) renamePath(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) deletePath(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) writeNewFile(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) makeDir(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) attachShell(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) resizeShell(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) runCommand(w http.ResponseWriter, r *http.Request) {
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
			"/containers/{id:.*}/attach":  s.attachShell,
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
	s.Router = s.CreateRouter()
	return s
}
