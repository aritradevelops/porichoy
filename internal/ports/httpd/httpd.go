package httpd

import (
	"fmt"

	"github.com/aritradeveops/porichoy/internal/config"
	"github.com/aritradeveops/porichoy/internal/pkg/translation"
	"github.com/aritradeveops/porichoy/internal/ports/httpd/handlers"
	"github.com/aritradeveops/porichoy/internal/ports/httpd/middlewares"
	"github.com/aritradeveops/porichoy/internal/ports/httpd/ui"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	app      *fiber.App
	config   *config.Config
	handlers *handlers.Handlers
	ui       *ui.UI
}

func NewServer(config *config.Config, handlers *handlers.Handlers, ui *ui.UI) *Server {
	app := fiber.New(fiber.Config{
		ErrorHandler: middlewares.ErrorHandler(),
	})
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	app.Use(middlewares.RequestLogger())
	app.Use(translation.New())
	server := &Server{
		config:   config,
		app:      app,
		handlers: handlers,
		ui:       ui,
	}
	server.setupRoutes()
	return server
}

func (s *Server) Start() error {
	return s.app.Listen(fmt.Sprintf("%s:%d", s.config.Http.Host, s.config.Http.Port))
}
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
