package ui

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"timber-ssh-go/ui/viewmodel"
)

type App struct {
	deps RouterDeps
}

type RouterDeps struct {
	HomeData      func(lang string) viewmodel.HomePageData
	DetailData    func(id string, lang string) (viewmodel.DetailPageData, bool)
	NewFormData   func(lang string) viewmodel.FormPageData
	EditFormData  func(id string, lang string) (viewmodel.FormPageData, bool)
	SaveProfile   func(r *http.Request) (string, error)
	StartProfile  func(id string) error
	StopProfile   func(id string) error
	DeleteProfile func(id string) error
}

func NewApp(deps RouterDeps) *App {
	return &App{deps: deps}
}

func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/", a.handleHome)
	r.Get("/profiles/new", a.handleNewProfile)
	r.Get("/profiles/{id}", a.handleDetail)
	r.Get("/profiles/{id}/diagnostics", a.handleDiagnostics)
	r.Get("/profiles/{id}/edit", a.handleEditProfile)
	r.Post("/profiles/save", a.handleSaveProfile)
	r.Post("/profiles/{id}/start", a.handleStartProfile)
	r.Post("/profiles/{id}/stop", a.handleStopProfile)
	r.Post("/profiles/{id}/delete", a.handleDeleteProfile)
	return r
}
