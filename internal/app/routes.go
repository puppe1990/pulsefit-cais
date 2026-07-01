package app

import (
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"github.com/puppe1990/pulsefit/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps) {
	r.Use(middleware.LoadSession(deps.SessionStore))

	auth := handlers.NewAuthHandler(deps.Renderer, deps.Store, deps.SessionStore, deps.SecureCookie)
	pages := handlers.NewPagesHandler(deps.Renderer, deps.Store)
	workout := handlers.NewWorkoutHandler(deps.Renderer, deps.Store)

	r.Get("/login", auth.Login)
	r.Post("/login", auth.LoginPost)
	r.Get("/register", auth.Register)
	r.Post("/register", auth.RegisterPost)
	r.Post("/logout", auth.Logout)

	r.Group(middleware.RequireAuth("/login"), func(g *cais.Router) {
		g.Get("/", pages.Dashboard)
		g.Get("/search", pages.Search)
		g.Get("/library", pages.Library)
		g.Get("/routine/new", pages.RoutineNew)
		g.Post("/routine/new", pages.RoutineCreatePost)
		g.Get("/routine/{id}", cais.StringParam("id", pages.RoutineDetail))
		g.Post("/routine/{id}/start", cais.StringParam("id", workout.Start))
		g.Get("/workout/{id}", cais.StringParam("id", workout.Show))
		g.Post("/workout/{id}/set", cais.StringParam("id", workout.AddSet))
		g.Post("/workout/{id}/finish", cais.StringParam("id", workout.Finish))
		g.Get("/workout/{id}/summary", cais.StringParam("id", workout.Summary))
		g.Get("/history", pages.History)
	})
}