package template

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed vanilla/*
var vanillaTemplate embed.FS

func GetTemplate(id string) (fs.FS, error) {
	switch id {
	case "vanilla":
		return fs.Sub(vanillaTemplate, "vanilla")
	default:
		return nil, fmt.Errorf("template not found")
	}
}
