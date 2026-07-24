package user

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/internal/webclient"
)

func TestLogin_HappyPath(t *testing.T) {
	var signInBody map[string]any
	var mintBody map[string]any
	var mintCSRF string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/sign_in":
			if r.Method == http.MethodGet {
				_, _ = io.WriteString(w, `<meta name="csrf-token" content="pre" />`)
				return
			}
			_ = json.NewDecoder(r.Body).Decode(&signInBody)
			http.SetCookie(w, &http.Cookie{Name: "_session", Value: "auth", Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
			w.Header().Set("Location", "/dashboard")
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{}`)
		case "/me/profile/security":
			_, _ = io.WriteString(w, `<meta name="csrf-token" content="post" />`)
		case "/me/profile/security/user_auth_tokens":
			mintCSRF = r.Header.Get("X-CSRF-Token")
			_ = json.NewDecoder(r.Body).Decode(&mintBody)
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"token":"bitpat_xyz","slug":"tok-1","api_url":"https://api.bitrise.io"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c, _ := webclient.New(srv.URL)
	svc := NewService(c)
	tok, err := svc.Login(context.Background(), LoginInput{Login: "alice@example.com", Password: "pw"}, "bitrise-cli (host)")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if tok != "bitpat_xyz" {
		t.Fatalf("token = %q, want bitpat_xyz", tok)
	}
	if got, ok := signInBody["user"].(map[string]any); !ok || got["login"] != "alice@example.com" || got["password"] != "pw" {
		t.Fatalf("sign_in body = %#v", signInBody)
	}
	if mintBody["description"] != "bitrise-cli (host)" {
		t.Fatalf("mint description = %v", mintBody["description"])
	}
	if mintBody["registration_type"] != "manual" {
		t.Fatalf("mint registration_type = %v, want manual", mintBody["registration_type"])
	}
	if mintCSRF != "post" {
		t.Fatalf("mint X-CSRF-Token = %q, want %q (re-prime after sign-in must refresh it)", mintCSRF, "post")
	}
}

func TestLogin_UnconfirmedEmail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/sign_in" && r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `<meta name="csrf-token" content="t" />`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"You have to confirm your email address before continuing."}`)
	}))
	defer srv.Close()

	c, _ := webclient.New(srv.URL)
	svc := NewService(c)
	_, err := svc.Login(context.Background(), LoginInput{Login: "a@b", Password: "p"}, "desc")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !IsUnconfirmedEmailErr(err) {
		t.Fatalf("IsUnconfirmedEmailErr(%v) = false, want true", err)
	}
}

func TestLogin_GenericFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/sign_in" && r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `<meta name="csrf-token" content="t" />`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"Invalid Email or password."}`)
	}))
	defer srv.Close()

	c, _ := webclient.New(srv.URL)
	svc := NewService(c)
	_, err := svc.Login(context.Background(), LoginInput{Login: "a@b", Password: "p"}, "desc")
	if err == nil {
		t.Fatalf("expected error")
	}
	if IsUnconfirmedEmailErr(err) {
		t.Fatalf("expected NOT unconfirmed; got unconfirmed sentinel")
	}
	if !strings.Contains(err.Error(), "Invalid Email or password") {
		t.Fatalf("error %q does not surface server message", err.Error())
	}
}

func TestLogin_FieldErrorsShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/sign_in" && r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `<meta name="csrf-token" content="t" />`)
			return
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, `{"errors":{"login":[{"error":"invalid"}],"password":[{"error":"too_short"}]}}`)
	}))
	defer srv.Close()

	c, _ := webclient.New(srv.URL)
	svc := NewService(c)
	_, err := svc.Login(context.Background(), LoginInput{Login: "a@b", Password: "p"}, "desc")
	if err == nil {
		t.Fatalf("expected error")
	}
	got := err.Error()
	if !strings.Contains(got, "login: invalid") || !strings.Contains(got, "password: too_short") || !strings.Contains(got, "422") {
		t.Fatalf("error %q does not include expected field details", got)
	}
}

func TestLogin_MintFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/users/sign_in" && r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `<meta name="csrf-token" content="t" />`)
		case r.URL.Path == "/users/sign_in":
			w.WriteHeader(http.StatusCreated)
		case r.URL.Path == "/me/profile/security":
			_, _ = io.WriteString(w, `<meta name="csrf-token" content="t2" />`)
		case r.URL.Path == "/me/profile/security/user_auth_tokens":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `boom`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c, _ := webclient.New(srv.URL)
	svc := NewService(c)
	_, err := svc.Login(context.Background(), LoginInput{Login: "a@b", Password: "p"}, "desc")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "mint access token failed") {
		t.Fatalf("unexpected error %q", err.Error())
	}
}
