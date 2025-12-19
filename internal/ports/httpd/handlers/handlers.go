package handlers

import (
	"github.com/aritradeveops/porichoy/internal/core/service"
	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	service *service.Service
}

func New(srv *service.Service) *Handlers {
	return &Handlers{
		service: srv,
	}
}

func (h *Handlers) Hello(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Hello!",
	})
}
