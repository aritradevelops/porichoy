package httpd

import (
	"net/http"

	"github.com/aritradeveops/porichoy/internal/ports/httpd/authn"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

func (s *Server) setupRoutes() {
	router := s.app
	router.Use("/", filesystem.New(filesystem.Config{
		Root: http.Dir("./template/vanilla"),
	}))
	apiRouter := router.Group("/api/v1")

	apiRouter.Get("/", s.handlers.Hello)

	authRouter := apiRouter.Group("/auth")
	authRouter.Post("/register", s.handlers.RegisterUser)
	authRouter.Post("/login", s.handlers.LoginUser)
	authRouter.Get("/oauth2", authn.Middleware(s.config.Authentication.JWT, true), s.handlers.Oauth2)
	appRouter := apiRouter.Group("/apps", authn.Middleware(s.config.Authentication.JWT))
	appRouter.Post("/create", s.handlers.CreateApp)
	configRouter := apiRouter.Group("/config")
	configRouter.Post("/configure", s.handlers.Configure)
}
