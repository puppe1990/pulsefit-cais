package handlers

import (
	"net/http"
	"strconv"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/session"

	"github.com/puppe1990/pulsefit/internal/models"
	"github.com/puppe1990/pulsefit/internal/store"
)

type AppLayoutData struct {
	meta.Site
	ActiveNav string
	Profile   Profile
}

type DashboardData struct {
	AppLayoutData
	FirstName string
	Routines  []RoutineCard
}

type SearchData struct {
	AppLayoutData
	Categories []SearchCategory
	Recent     []string
}

type LibraryData struct {
	AppLayoutData
	Exercises []ExerciseCard
}

type RoutineExerciseRow struct {
	Position int
	ExerciseCard
}

type RoutineDetailData struct {
	AppLayoutData
	Routine   RoutineCard
	Exercises []RoutineExerciseRow
}

type RoutineNewData struct {
	AppLayoutData
	Exercises []ExerciseCard
}

type HistoryData struct {
	AppLayoutData
	Sessions []HistorySession
}

type PagesHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
}

func NewPagesHandler(renderer *cais.Renderer, st store.Store, site meta.Site) *PagesHandler {
	return &PagesHandler{renderer: renderer, store: st, site: site}
}

func (h *PagesHandler) layout(r *http.Request, active string) AppLayoutData {
	profile := Profile{DisplayName: "Athlete", PhotoURL: "https://api.dicebear.com/7.x/avataaars/svg?seed=guest"}
	if id, ok := session.UserID(r); ok {
		if u, err := h.store.FindUserByID(id); err == nil {
			profile = Profile{DisplayName: u.DisplayName, PhotoURL: u.PhotoURL, Email: u.Email}
		}
	}
	return AppLayoutData{Site: h.site, ActiveNav: active, Profile: profile}
}

func (h *PagesHandler) render(w http.ResponseWriter, layout, page string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, layout, page, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *PagesHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	userID, _ := session.UserID(r)
	layout := h.layout(r, "home")

	routines, err := h.store.ListRoutinesByUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cards := make([]RoutineCard, len(routines))
	for i, rt := range routines {
		cards[i] = routineCard(rt)
	}

	h.render(w, "base", "dashboard", DashboardData{
		AppLayoutData: layout,
		FirstName:     firstName(layout.Profile.DisplayName),
		Routines:      cards,
	})
}

func (h *PagesHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID, _ := session.UserID(r)

	exercises, err := h.store.ListExercises()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	routines, err := h.store.ListRoutinesByUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	recent := make([]string, 0, 3)
	for i, rt := range routines {
		if i >= 3 {
			break
		}
		recent = append(recent, rt.Name)
	}

	h.render(w, "base", "search", SearchData{
		AppLayoutData: h.layout(r, "search"),
		Categories:    categoriesFromExercises(exercises),
		Recent:        recent,
	})
}

func (h *PagesHandler) Library(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.store.ListExercises()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cards := make([]ExerciseCard, len(exercises))
	for i, ex := range exercises {
		cards[i] = exerciseCard(ex)
	}

	h.render(w, "base", "library", LibraryData{
		AppLayoutData: h.layout(r, "library"),
		Exercises:     cards,
	})
}

func (h *PagesHandler) RoutineNew(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.store.ListExercises()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cards := make([]ExerciseCard, len(exercises))
	for i, ex := range exercises {
		cards[i] = exerciseCard(ex)
	}

	h.render(w, "base", "routine_new", RoutineNewData{
		AppLayoutData: h.layout(r, "routine_new"),
		Exercises:     cards,
	})
}

func (h *PagesHandler) RoutineDetail(w http.ResponseWriter, r *http.Request, id string) {
	routineID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || routineID <= 0 {
		http.NotFound(w, r)
		return
	}

	userID, _ := session.UserID(r)
	routine, err := h.store.FindRoutineByID(routineID)
	if err != nil || routine.UserID != userID {
		http.NotFound(w, r)
		return
	}

	exercises, err := h.store.ListRoutineExercises(routineID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows := make([]RoutineExerciseRow, len(exercises))
	for i, ex := range exercises {
		rows[i] = RoutineExerciseRow{Position: i + 1, ExerciseCard: exerciseCard(ex)}
	}

	h.render(w, "base", "routine_detail", RoutineDetailData{
		AppLayoutData: h.layout(r, ""),
		Routine:       routineCard(routine),
		Exercises:     rows,
	})
}

func (h *PagesHandler) RoutineCreatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	var exerciseIDs []int64
	for _, raw := range r.Form["exercise_id"] {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			continue
		}
		exerciseIDs = append(exerciseIDs, id)
	}

	routineID, err := h.store.CreateRoutine(userID, models.Routine{
		Name:        name,
		Description: r.FormValue("description"),
		Emoji:       "🏋️",
		Color:       "bg-blue-600",
	}, exerciseIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.SeeOther(w, r, "/routine/"+strconv.FormatInt(routineID, 10))
}

func (h *PagesHandler) History(w http.ResponseWriter, r *http.Request) {
	userID, _ := session.UserID(r)

	sessions, err := h.store.ListSessionsByUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items := make([]HistorySession, len(sessions))
	for i, sess := range sessions {
		items[i] = historySession(sess)
	}

	h.render(w, "base", "history", HistoryData{
		AppLayoutData: h.layout(r, "history"),
		Sessions:      items,
	})
}
