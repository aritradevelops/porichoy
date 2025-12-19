package httpd

import (
	"github.com/aritradeveops/porichoy/internal/ports/httpd/authn"
)

func (s *Server) setupRoutes() {
	router := s.app

	apiRouter := router.Group("/api/v1")

	apiRouter.Get("/", s.handlers.Hello)

	authRouter := apiRouter.Group("/auth")
	authRouter.Post("/register", s.handlers.RegisterUser)
	authRouter.Post("/login", s.handlers.LoginUser)
	authRouter.Get("/profile", authn.Middleware(s.config.Authentication.JWT), s.handlers.GetUserProfile)
	appRouter := apiRouter.Group("/apps", authn.Middleware(s.config.Authentication.JWT))
	appRouter.Post("/create", s.handlers.CreateApp)

}
