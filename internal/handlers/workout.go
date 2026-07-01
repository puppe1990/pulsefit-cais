package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/session"
	"github.com/puppe1990/pulsefit/internal/store"
)

var errWorkoutNotFound = errors.New("workout not found")

type WorkoutHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
}

type WorkoutQueueItem struct {
	Position int
	Name     string
	Active   bool
	Done     bool
	URL      string
}

type WorkoutPageData struct {
	AppLayoutData
	SessionID     string
	RoutineName   string
	ExerciseIndex int
	ExerciseCount int
	DisplayIndex  int
	PrevURL       string
	NextURL       string
	ExerciseName  string
	ExerciseLogID int64
	ImageURL      string
	Instructions  string
	Queue         []WorkoutQueueItem
	Sets          []SetRow
	Elapsed       string
}

type SetRow struct {
	ID        int64
	Position  int
	Weight    float64
	Reps      int
	Completed bool
}

type WorkoutSummaryData struct {
	AppLayoutData
	RoutineName   string
	Duration      string
	ExerciseCount int
	TotalSets     int
	CompletedSets int
}

func NewWorkoutHandler(renderer *cais.Renderer, st store.Store, site meta.Site) *WorkoutHandler {
	return &WorkoutHandler{renderer: renderer, store: st, site: site}
}

func (h *WorkoutHandler) Start(w http.ResponseWriter, r *http.Request, routineID string) {
	userID, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	rid, err := strconv.ParseInt(routineID, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	sessionID, err := h.store.StartWorkout(userID, rid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	httpx.SeeOther(w, r, "/workout/"+strconv.FormatInt(sessionID, 10))
}

func (h *WorkoutHandler) Show(w http.ResponseWriter, r *http.Request, sessionID string) {
	data, err := h.buildWorkoutPage(r, sessionID)
	if err != nil {
		if errors.Is(err, errWorkoutNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderPage(w, "workout", data)
}

func (h *WorkoutHandler) AddSet(w http.ResponseWriter, r *http.Request, sessionID string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	data, err := h.buildWorkoutPage(r, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logID, _ := strconv.ParseInt(r.FormValue("exercise_log_id"), 10, 64)
	weight, _ := strconv.ParseFloat(r.FormValue("weight"), 64)
	reps, _ := strconv.Atoi(r.FormValue("reps"))
	completed := r.FormValue("completed") == "1"

	if _, err := h.store.AddSet(logID, weight, reps, completed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = h.buildWorkoutPage(r, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if cais.IsHTMX(r) {
		h.renderPartial(w, "workout_sets", struct {
			Sets          []SetRow
			ExerciseLogID int64
			SessionID     string
			ExerciseIndex int
		}{
			Sets: data.Sets, ExerciseLogID: data.ExerciseLogID, SessionID: sessionID, ExerciseIndex: data.ExerciseIndex,
		})
		return
	}
	h.renderPage(w, "workout", data)
}

func (h *WorkoutHandler) Finish(w http.ResponseWriter, r *http.Request, sessionID string) {
	userID, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	sid, err := strconv.ParseInt(sessionID, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sess, err := h.store.FindWorkoutSession(sid)
	if err != nil || sess.UserID != userID {
		http.NotFound(w, r)
		return
	}
	if err := h.store.FinishWorkout(sid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	httpx.SeeOther(w, r, "/workout/"+sessionID+"/summary")
}

func (h *WorkoutHandler) Summary(w http.ResponseWriter, r *http.Request, sessionID string) {
	userID, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	sid, err := strconv.ParseInt(sessionID, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	summary, err := h.store.WorkoutSummary(sid)
	if err != nil || summary.Session.UserID != userID {
		http.NotFound(w, r)
		return
	}
	logs, _ := h.store.ListExerciseLogs(sid)

	pages := NewPagesHandler(h.renderer, h.store, h.site)
	h.renderPage(w, "workout_summary", WorkoutSummaryData{
		AppLayoutData: pages.layout(r, ""),
		RoutineName:   summary.Session.RoutineName,
		Duration:      formatDuration(summary.Session.DurationSeconds),
		ExerciseCount: len(logs),
		TotalSets:     summary.TotalSets,
		CompletedSets: summary.CompletedSets,
	})
}

func (h *WorkoutHandler) buildWorkoutPage(r *http.Request, sessionID string) (WorkoutPageData, error) {
	userID, ok := session.UserID(r)
	if !ok {
		return WorkoutPageData{}, errWorkoutNotFound
	}
	sid, err := strconv.ParseInt(sessionID, 10, 64)
	if err != nil {
		return WorkoutPageData{}, errWorkoutNotFound
	}
	sess, err := h.store.FindWorkoutSession(sid)
	if err != nil || sess.UserID != userID {
		return WorkoutPageData{}, errWorkoutNotFound
	}
	if sess.CompletedAt != nil {
		return WorkoutPageData{}, errWorkoutNotFound
	}

	logs, err := h.store.ListExerciseLogs(sid)
	if err != nil {
		return WorkoutPageData{}, err
	}
	exIndex := 0
	if q := r.URL.Query().Get("ex"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n >= 0 && n < len(logs) {
			exIndex = n
		}
	}
	current := logs[exIndex]

	var imageURL, instructions string
	if current.ExerciseID > 0 {
		if ex, err := h.store.FindExerciseByID(current.ExerciseID); err == nil {
			imageURL = ex.ImageURL
			instructions = ex.Instructions
		}
	}

	sets, err := h.store.ListSets(current.ID)
	if err != nil {
		return WorkoutPageData{}, err
	}
	setRows := make([]SetRow, len(sets))
	for i, s := range sets {
		setRows[i] = SetRow{ID: s.ID, Position: i + 1, Weight: s.Weight, Reps: s.Reps, Completed: s.Completed}
	}

	queue := make([]WorkoutQueueItem, len(logs))
	for i, log := range logs {
		done := false
		if ss, err := h.store.ListSets(log.ID); err == nil {
			for _, s := range ss {
				if s.Completed {
					done = true
					break
				}
			}
		}
		queue[i] = WorkoutQueueItem{
			Position: i + 1,
			Name:     log.ExerciseName,
			Active:   i == exIndex,
			Done:     done,
			URL:      "/workout/" + sessionID + "?ex=" + strconv.Itoa(i),
		}
	}

	elapsed := formatElapsed(int(time.Since(sess.StartedAt).Seconds()))

	prevURL, nextURL := "", ""
	if exIndex > 0 {
		prevURL = "/workout/" + sessionID + "?ex=" + strconv.Itoa(exIndex-1)
	}
	if exIndex < len(logs)-1 {
		nextURL = "/workout/" + sessionID + "?ex=" + strconv.Itoa(exIndex+1)
	}

	pages := NewPagesHandler(h.renderer, h.store, h.site)
	return WorkoutPageData{
		AppLayoutData: pages.layout(r, ""),
		SessionID:     sessionID,
		RoutineName:   sess.RoutineName,
		ExerciseIndex: exIndex,
		ExerciseCount: len(logs),
		DisplayIndex:  exIndex + 1,
		PrevURL:       prevURL,
		NextURL:       nextURL,
		ExerciseName:  current.ExerciseName,
		ExerciseLogID: current.ID,
		ImageURL:      imageURL,
		Instructions:  instructions,
		Queue:         queue,
		Sets:          setRows,
		Elapsed:       elapsed,
	}, nil
}

func (h *WorkoutHandler) renderPage(w http.ResponseWriter, page string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, "base", page, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WorkoutHandler) renderPartial(w http.ResponseWriter, partial string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.RenderPartial(w, partial, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}