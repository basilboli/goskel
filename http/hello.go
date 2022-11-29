package http

import (
	"log"
	"net/http"
)

func (s *Server) handleHello(w http.ResponseWriter, r *http.Request) {
	log.Println("Hello request")
	// return ok
	w.WriteHeader(http.StatusOK)
}
