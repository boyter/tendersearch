package app

import (
	"log/slog"
	"net/http"

	"github.com/boyter/tendersearch/internal/common"
)

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	slog.Info("home", common.UC("e1007e12"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := homeTemplate.Execute(w, templateData{})
	if err != nil {
		slog.Error("error", "err", err, common.UC("6a4447c5"))
	}
}
