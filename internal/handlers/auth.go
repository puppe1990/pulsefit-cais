package handlers

import (
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/session"
	"github.com/puppe1990/pulsefit/internal/models"
	"github.com/puppe1990/pulsefit/internal/store"
)

type AuthHandler struct {
	renderer *cais.Renderer
	store    store.Store
	sessions session.Store
	secure   bool
}

type LoginPageData struct {
	Error string
}

type RegisterPageData struct {
	Error string
}

func NewAuthHandler(renderer *cais.Renderer, st store.Store, sessions session.Store, secure bool) *AuthHandler {
	return &AuthHandler{renderer: renderer, store: st, sessions: sessions, secure: secure}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.renderAuth(w, "login", LoginPageData{})
}

func (h *AuthHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	u, err := h.store.FindUserByEmail(email)
	if err != nil || !session.VerifyPassword(u.PasswordHash, password) {
		h.renderAuth(w, "login", LoginPageData{Error: "Invalid email or password"})
		return
	}

	if err := session.SignIn(w, h.sessions, u.ID, session.CookieOptions{Secure: h.secure}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.SeeOther(w, r, "/")
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.renderAuth(w, "register", RegisterPageData{})
}

func (h *AuthHandler) RegisterPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	if displayName == "" {
		displayName = strings.Split(email, "@")[0]
	}

	if email == "" || password == "" {
		h.renderAuth(w, "register", RegisterPageData{Error: "Email and password are required"})
		return
	}

	hash, err := session.HashPassword(password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := h.store.CreateUser(models.User{
		Email:        email,
		PasswordHash: hash,
		DisplayName:  displayName,
		PhotoURL:     "https://api.dicebear.com/7.x/avataaars/svg?seed=" + email,
	})
	if err != nil {
		h.renderAuth(w, "register", RegisterPageData{Error: "Could not create account — email may already exist"})
		return
	}

	if err := session.SignIn(w, h.sessions, id, session.CookieOptions{Secure: h.secure}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.SeeOther(w, r, "/")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session.SignOut(w, h.sessions, r)
	httpx.SeeOther(w, r, "/login")
}

func (h *AuthHandler) renderAuth(w http.ResponseWriter, page string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, "auth", page, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}