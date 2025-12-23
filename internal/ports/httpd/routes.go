package httpd

import (
	"github.com/aritradeveops/porichoy/internal/ports/httpd/authn"
)

func (s *Server) setupRoutes() {
	router := s.app
	// router.Use("/", filesystem.New(filesystem.Config{
	// 	Root: http.Dir("./template/vanilla"),
	// }))

	router.Get("/", s.ui.Index)
	router.Get("/login", s.ui.Login)
	router.Get("/register", s.ui.Register)

	apiRouter := router.Group("/api/v1")
	apiRouter.Get("/", s.handlers.Hello)

	authRouter := apiRouter.Group("/auth")
	authRouter.Post("/register", s.handlers.RegisterUser)
	authRouter.Post("/login", s.handlers.LoginUser)
	authRouter.Get("/oauth2", authn.Middleware(true), s.handlers.Oauth2)
	authRouter.Post("/token", s.handlers.Token)
	appRouter := apiRouter.Group("/apps", authn.Middleware())
	appRouter.Post("/create", s.handlers.CreateApp)
	configRouter := apiRouter.Group("/config")
	configRouter.Post("/configure", s.handlers.Configure)
}
