package http

func (s *Server) routes() {
	s.router.Get("/", s.handleVersion)
	s.router.Get("/health", s.handleHealthCheck)
	s.router.Get("/hello", s.handleHello)

	// this endpoint is dedicated to cron jobs
	s.router.Post("/jobs", s.adminOnly(s.handleS3ExportJob()))
}
