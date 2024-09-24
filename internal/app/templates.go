package app

import (
	"html/template"

	"github.com/boyter/tendersearch/assets"
)

// Full page templates
var homeTemplate *template.Template

// For passing into templates
type templateData struct{}

func (app *Application) ParseTemplates() error {
	t, err := template.ParseFS(assets.Assets, "public/html/home.html")
	if err != nil {
		return err
	}
	homeTemplate = t

	return nil
}
