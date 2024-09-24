package app

import (
	"net/http"

	"github.com/boyter/tendersearch/assets"
	"github.com/boyter/tendersearch/internal/service"
	"github.com/gorilla/mux"
)

type Application struct {
	Service *service.Service
}

func NewApplication(ser *service.Service) Application {
	return Application{
		Service: ser,
	}
}

func (app *Application) Router() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", app.Home).Methods("GET")
	router.HandleFunc("/v1/health-check/", app.HealthCheck).Methods("GET")

	addContentRoutes(router)

	return router
}

func addContentRoutes(router *mux.Router) {
	router.HandleFunc("/style.css", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write([]byte(assets.Css))
	})

	router.HandleFunc("/logo.png", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(assets.Logo)
	})
}
