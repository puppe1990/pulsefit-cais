package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais/session"

	"github.com/puppe1990/pulsefit/internal/store"
)

func setupAuthHandler(t *testing.T) (*AuthHandler, *store.SQLiteStore) {
	t.Helper()
	st := setupTestSQLiteStore(t)
	sessions := store.NewSessionStore(st)
	return NewAuthHandler(setupTestRenderer(t), st, sessions, false, testSite()), st
}

func TestAuthHandler_LoginPost_success(t *testing.T) {
	h, _ := setupAuthHandler(t)

	form := url.Values{"email": {"demo@pulsefit.local"}, "password": {"demo"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.LoginPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/" {
		t.Errorf("Location = %q, want /", loc)
	}
	cookies := rr.Result().Cookies()
	defer func() { _ = rr.Result().Body.Close() }()
	found := false
	for _, c := range cookies {
		if c.Name == session.DefaultCookieName && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected session cookie")
	}
}

func TestAuthHandler_LoginPost_invalidPassword(t *testing.T) {
	h, _ := setupAuthHandler(t)

	form := url.Values{"email": {"demo@pulsefit.local"}, "password": {"wrong"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.LoginPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid email or password") {
		t.Error("expected error message")
	}
}

func TestAuthHandler_RegisterPost_createsUser(t *testing.T) {
	h, st := setupAuthHandler(t)

	form := url.Values{
		"email":        {"new@example.com"},
		"password":     {"secret123"},
		"display_name": {"New Athlete"},
	}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.RegisterPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	u, err := st.FindUserByEmail("new@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if u.DisplayName != "New Athlete" {
		t.Errorf("display name = %q", u.DisplayName)
	}
}

func TestAuthHandler_Logout_clearsSession(t *testing.T) {
	h, st := setupAuthHandler(t)
	sessions := store.NewSessionStore(st)
	u, err := st.FindUserByEmail("demo@pulsefit.local")
	if err != nil {
		t.Fatal(err)
	}
	token, err := sessions.Create(u.ID)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: session.DefaultCookieName, Value: token})
	rr := httptest.NewRecorder()
	h.Logout(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	if _, ok := sessions.Get(token); ok {
		t.Error("session should be deleted")
	}
}
