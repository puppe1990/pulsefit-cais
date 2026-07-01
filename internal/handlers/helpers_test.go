package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/session"
	"github.com/puppe1990/cais/pkg/cais/testutil"
	"github.com/puppe1990/pulsefit/internal/store"
)

func testSite() meta.Site {
	return meta.Site{AppName: "PulseFit", AppURL: "https://pulsefit.gestaobem.com"}
}

func setupTestRenderer(t *testing.T) *cais.Renderer {
	t.Helper()
	return testutil.NewRenderer(t)
}

func setupTestStore(t *testing.T) store.Store {
	t.Helper()
	return setupTestSQLiteStore(t)
}

func setupTestSQLiteStore(t *testing.T) *store.SQLiteStore {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if err := s.SeedDemo(); err != nil {
		t.Fatal(err)
	}
	return s
}

func demoUserID(t *testing.T, s store.Store) int64 {
	t.Helper()
	u, err := s.FindUserByEmail("demo@pulsefit.local")
	if err != nil {
		t.Fatal(err)
	}
	return u.ID
}

func authedRequest(t *testing.T, s store.Store, method, path string, body ...io.Reader) *http.Request {
	t.Helper()
	var r io.Reader
	if len(body) > 0 {
		r = body[0]
	}
	req := httptest.NewRequest(method, path, r)
	return session.WithUserID(req, demoUserID(t, s))
}
