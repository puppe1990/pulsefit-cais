package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestWorkoutHandler_Start_redirects(t *testing.T) {
	st := setupTestStore(t)
	h := NewWorkoutHandler(setupTestRenderer(t), st, testSite())
	uID := demoUserID(t, st)
	routines, _ := st.ListRoutinesByUser(uID)
	rid := strconv.FormatInt(routines[0].ID, 10)

	rr := httptest.NewRecorder()
	h.Start(rr, authedRequest(t, st, http.MethodPost, "/routine/"+rid+"/start"), rid)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	if !strings.HasPrefix(rr.Header().Get("Location"), "/workout/") {
		t.Errorf("location = %q", rr.Header().Get("Location"))
	}
}

func TestWorkoutHandler_Show_andAddSet(t *testing.T) {
	st := setupTestStore(t)
	h := NewWorkoutHandler(setupTestRenderer(t), st, testSite())
	uID := demoUserID(t, st)
	routines, _ := st.ListRoutinesByUser(uID)
	sessionID, err := st.StartWorkout(uID, routines[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	sid := strconv.FormatInt(sessionID, 10)

	rr := httptest.NewRecorder()
	h.Show(rr, authedRequest(t, st, http.MethodGet, "/workout/"+sid), sid)
	if rr.Code != http.StatusOK {
		t.Fatalf("show status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Log set") {
		t.Error("missing set form")
	}

	logs, _ := st.ListExerciseLogs(sessionID)
	form := url.Values{
		"exercise_log_id": {strconv.FormatInt(logs[0].ID, 10)},
		"weight":          {"60"},
		"reps":            {"10"},
		"completed":       {"1"},
	}
	req := authedRequest(t, st, http.MethodPost, "/workout/"+sid+"/set?ex=0", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr = httptest.NewRecorder()
	h.AddSet(rr, req, sid)
	if rr.Code != http.StatusOK {
		t.Fatalf("add set status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "60") {
		t.Error("missing logged set in partial")
	}
}

func TestWorkoutHandler_Finish_andSummary(t *testing.T) {
	st := setupTestStore(t)
	h := NewWorkoutHandler(setupTestRenderer(t), st, testSite())
	uID := demoUserID(t, st)
	routines, _ := st.ListRoutinesByUser(uID)
	sessionID, _ := st.StartWorkout(uID, routines[0].ID)
	sid := strconv.FormatInt(sessionID, 10)

	rr := httptest.NewRecorder()
	h.Finish(rr, authedRequest(t, st, http.MethodPost, "/workout/"+sid+"/finish"), sid)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("finish status = %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	h.Summary(rr, authedRequest(t, st, http.MethodGet, "/workout/"+sid+"/summary"), sid)
	if rr.Code != http.StatusOK {
		t.Fatalf("summary status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Workout") {
		t.Error("missing summary heading")
	}
}

func TestPagesHandler_RoutineCreatePost(t *testing.T) {
	st := setupTestStore(t)
	h := NewPagesHandler(setupTestRenderer(t), st, testSite())
	exercises, _ := st.ListExercises()

	form := url.Values{"name": {"My Test Routine"}, "exercise_id": {strconv.FormatInt(exercises[0].ID, 10)}}
	req := authedRequest(t, st, http.MethodPost, "/routine/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.RoutineCreatePost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.HasPrefix(rr.Header().Get("Location"), "/routine/") {
		t.Errorf("location = %q", rr.Header().Get("Location"))
	}
}