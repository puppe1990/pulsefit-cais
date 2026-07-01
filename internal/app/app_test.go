package app

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/pulsefit/internal/store"
	"github.com/puppe1990/pulsefit/web"
)

func TestApp_StaticCSS_PublicWithoutAuth(t *testing.T) {
	t.Chdir(findAppRoot(t))

	tmplFS, err := fs.Sub(web.Templates, "templates")
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := cais.NewRenderer(tmplFS)
	if err != nil {
		t.Fatal(err)
	}

	st, err := store.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.SeedDemo(); err != nil {
		t.Fatal(err)
	}

	staticDir := filepath.Join("web", "static")
	a, err := New(cais.Config{Port: ":0", DBPath: ":memory:"}, Deps{
		Renderer:     renderer,
		Store:        st,
		SessionStore: store.NewSessionStore(st),
		StaticDir:    staticDir,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/static/css/styles.css", nil)
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/css", ct)
	}
}

func findAppRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(wd, "web", "static")); err == nil {
				return wd
			}
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("app root not found")
		}
		wd = parent
	}
}
