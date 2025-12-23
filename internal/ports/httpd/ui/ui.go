package ui

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type UI struct {
	template string
}

func New(template string) *UI {
	return &UI{
		template: template,
	}
}

func (u *UI) Index(c *fiber.Ctx) error {
	return c.SendFile(fmt.Sprintf("./template/%s/index.html", u.template))
}

func (u *UI) Login(c *fiber.Ctx) error {
	return c.SendFile(fmt.Sprintf("./template/%s/login.html", u.template))
}

func (u *UI) Register(c *fiber.Ctx) error {
	return c.SendFile(fmt.Sprintf("./template/%s/register.html", u.template))
}
