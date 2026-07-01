package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestPagesHandler_AllRoutes(t *testing.T) {
	st := setupTestStore(t)
	h := NewPagesHandler(setupTestRenderer(t), st, testSite())
	tests := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		path string
		want string
	}{
		{"dashboard", h.Dashboard, "/", "Welcome back"},
		{"search", h.Search, "/search", "Browse all"},
		{"library", h.Library, "/library", "Exercise"},
		{"routine_new", h.RoutineNew, "/routine/new", "Create Routine"},
		{"history", h.History, "/history", "Your"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tc.fn(rr, authedRequest(t, st, http.MethodGet, tc.path))
			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d", rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tc.want) {
				t.Errorf("body missing %q", tc.want)
			}
		})
	}
}

func TestPagesHandler_RoutineDetail(t *testing.T) {
	st := setupTestStore(t)
	h := NewPagesHandler(setupTestRenderer(t), st, testSite())

	uID := demoUserID(t, st)
	routines, err := st.ListRoutinesByUser(uID)
	if err != nil || len(routines) == 0 {
		t.Fatal("expected demo routines")
	}
	id := strconv.FormatInt(routines[0].ID, 10)

	rr := httptest.NewRecorder()
	h.RoutineDetail(rr, authedRequest(t, st, http.MethodGet, "/routine/"+id), id)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), routines[0].Name) {
		t.Error("missing routine name")
	}
}

func TestPagesHandler_Dashboard_showsDemoRoutines(t *testing.T) {
	st := setupTestStore(t)
	h := NewPagesHandler(setupTestRenderer(t), st, testSite())

	rr := httptest.NewRecorder()
	h.Dashboard(rr, authedRequest(t, st, http.MethodGet, "/"))
	if !strings.Contains(rr.Body.String(), "Upper Body Power") {
		t.Error("missing seeded routine")
	}
}