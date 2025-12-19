package translation

import (
	"os"
	"path"

	"github.com/aritradeveops/porichoy/internal/pkg/logger"
	"github.com/gofiber/contrib/fiberi18n/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.yaml.in/yaml/v3"
	"golang.org/x/text/language"
)

func New() fiber.Handler {
	cwd, _ := os.Getwd()
	localesPath := path.Join(cwd, "locales")
	return fiberi18n.New(&fiberi18n.Config{
		RootPath:         localesPath,
		AcceptLanguages:  []language.Tag{language.English, language.Bengali},
		DefaultLanguage:  language.English,
		UnmarshalFunc:    yaml.Unmarshal,
		FormatBundleFile: "yaml",
	})
}

func Localize(c *fiber.Ctx, id string, data ...any) string {
	var tmplData any
	if len(data) > 0 {
		tmplData = data[0]
	}
	msg, err := fiberi18n.Localize(c, &i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: tmplData,
	})
	if err != nil {
		logger.Warn().Str("key", id).Msg("Missing entry in locales")
		return id
	}
	return msg
}
