package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"goskel/service"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"github.com/go-chi/chi"
)

type Server struct {
	Service *service.Service
	router  *chi.Mux
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func NewServer() *Server {
	s := &Server{}
	s.router = chi.NewRouter()

	// A good base middleware stack goes here
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Heartbeat("/ping"))

	// Adding CORS middleware
	_cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	s.router.Use(_cors.Handler)

	// Add Content-Type application/json
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})

	s.routes()
	return s
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) error {
	w.WriteHeader(status)
	if data != nil {
		return json.NewEncoder(w).Encode(data)
	}
	return nil
}

// e.g. http.HandleFunc("/health-check", handleHealthCheck)
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our Service, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	_, _ = io.WriteString(w, `{"alive": true}`)
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(fmt.Sprintf("Commit hash: %s\nBuild time: %s", s.Service.Opts.CommitHash, s.Service.Opts.BuildTime)))
	if err != nil {
		log.Printf("[WARN] Problem writing response, %s", err)
	}
}

func (s *Server) handleS3ExportJob() http.HandlerFunc {

	type request struct {
		ConfigurationUuid string `json:"configurationUuid"`
		Format            string `json:"format"`
		ReferenceDate     string `json:"referenceDate"`
		ForceFileName     string `json:"forceFileName"`
	}

	type response struct {
		Status   string `json:"status"`
		FileName string `json:"filename"`
		Checksum string `json:"md5"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req *request
		err := decodeJSONBody(w, r, &req)
		if err != nil {
			sentry.CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if req.Format != "csv" {
			sentry.CaptureException(err)
			http.Error(w, "Not supported format. Currently supported formats : csv", http.StatusBadRequest)
			return
		}

		filename, checksum, err := s.Service.S3ExportJob(req.ConfigurationUuid)
		if err != nil {
			sentry.CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := &response{
			Status:   "ok",
			FileName: filename,
			Checksum: checksum,
		}

		err = s.respond(w, r, resp, http.StatusOK)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func (s *Server) adminOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// get token from Authorization header
		adminToken := r.Header.Get("Authorization")

		if adminToken == "" || adminToken != s.Service.Opts.AdminToken {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		// put UUID in context
		ctx := context.WithValue(r.Context(), "isAdmin", true)

		h(w, r.WithContext(ctx))
	}
}
