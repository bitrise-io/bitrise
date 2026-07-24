// Package user holds email/password sign-in against app.bitrise.io's
// Rails-Devise JSON endpoints (/users/sign_in for sign-in,
// /me/profile/security/user_auth_tokens to mint a PAT). Only the minted PAT
// is persisted, by the cli layer.
package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bitrise-io/bitrise/v2/internal/webclient"
)

// Service runs the login flow against an app.bitrise.io target.
type Service struct {
	client *webclient.Client
}

func NewService(client *webclient.Client) *Service {
	return &Service{client: client}
}

// LoginInput is the email/username + password payload. The wire field is
// "login" (Devise authentication_keys = [:login]); it accepts either an
// email or a username.
type LoginInput struct {
	Login    string
	Password string
}

// errUnconfirmedEmail is returned by Login when the server rejects
// credentials for an unconfirmed email.
var errUnconfirmedEmail = errors.New("email not yet verified")

// IsUnconfirmedEmailErr reports whether err is the unconfirmed-email
// sentinel returned by Login.
func IsUnconfirmedEmailErr(err error) bool { return errors.Is(err, errUnconfirmedEmail) }

// Login signs in via POST /users/sign_in and mints a PAT via
// POST /me/profile/security/user_auth_tokens. On an unconfirmed-email 401,
// returns an error satisfying IsUnconfirmedEmailErr.
func (s *Service) Login(ctx context.Context, in LoginInput, tokenDescription string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("webclient not configured")
	}
	// Sign-in itself skips CSRF on the server (sessions_controller.rb
	// `skip_before_action :verify_authenticity_token, only: [:create]`),
	// but we still prime so the session cookie is set and so the
	// subsequent PAT mint has a valid CSRF token.
	if err := s.client.Prime(ctx, "/users/sign_in"); err != nil {
		return "", fmt.Errorf("prime sign_in: %w", err)
	}
	signInBody := map[string]any{
		"user": map[string]string{
			"login":    in.Login,
			"password": in.Password,
		},
	}
	signIn, err := s.client.PostJSON(ctx, "/users/sign_in", signInBody)
	if err != nil {
		return "", err
	}
	if signIn.Status == http.StatusUnauthorized && looksLikeUnconfirmed(signIn.Body) {
		return "", errUnconfirmedEmail
	}
	if signIn.Status < 200 || signIn.Status >= 300 {
		return "", fmt.Errorf("sign in failed: %s", formatServerError(signIn.Status, signIn.Body))
	}

	// Devise rotates the CSRF token on successful authentication
	// (clean_up_csrf_token_on_authentication is true by default), so
	// the token captured before sign-in is stale. Re-prime an
	// authenticated page to pick up the fresh token before the mint
	// POST — without this the website's protect_from_forgery
	// raises InvalidAuthenticityToken → 422 (Unprocessable Content).
	if err := s.client.Prime(ctx, "/me/profile/security"); err != nil {
		return "", fmt.Errorf("re-prime after sign-in: %w", err)
	}

	// registration_type must be one of %w[manual login] (UserAuthToken
	// model). The controller forwards a nil param straight through to
	// the create service, which then trips inclusion validation → 422.
	// "manual" matches what the dashboard's "Create new token" UI sends.
	mintBody := map[string]any{
		"description":       tokenDescription,
		"registration_type": "manual",
	}
	mint, err := s.client.PostJSON(ctx, "/me/profile/security/user_auth_tokens", mintBody)
	if err != nil {
		return "", err
	}
	if mint.Status < 200 || mint.Status >= 300 {
		return "", fmt.Errorf("mint access token failed: %s", formatServerError(mint.Status, mint.Body))
	}
	var minted struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(mint.Body, &minted); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if minted.Token == "" {
		return "", fmt.Errorf("server returned an empty token")
	}
	return minted.Token, nil
}

// formatServerError pulls a human-readable phrase out of the website's
// JSON error envelope. Devise typically returns one of:
//
//	{"status":422,"error":"Unprocessable Content"}
//	{"errors":{"email":[{"error":"taken"}]}}
//	{"error":"Invalid Email or password."}
//
// We try each shape in turn and fall back to the raw body so nothing is
// silently swallowed.
func formatServerError(status int, body []byte) string {
	type errorsMap map[string][]map[string]any
	var envelope struct {
		Error  string    `json:"error"`
		Errors errorsMap `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil {
		if len(envelope.Errors) > 0 {
			parts := make([]string, 0, len(envelope.Errors))
			for field, details := range envelope.Errors {
				codes := make([]string, 0, len(details))
				for _, d := range details {
					if e, ok := d["error"].(string); ok && e != "" {
						codes = append(codes, e)
					}
				}
				if len(codes) == 0 {
					parts = append(parts, field)
				} else {
					parts = append(parts, fmt.Sprintf("%s: %s", field, strings.Join(codes, ", ")))
				}
			}
			return fmt.Sprintf("HTTP %d (%s)", status, strings.Join(parts, "; "))
		}
		if envelope.Error != "" {
			return fmt.Sprintf("HTTP %d (%s)", status, envelope.Error)
		}
	}
	if len(body) == 0 {
		return fmt.Sprintf("HTTP %d", status)
	}
	return fmt.Sprintf("HTTP %d: %s", status, strings.TrimSpace(string(body)))
}

// looksLikeUnconfirmed checks whether the server's error body matches
// Devise's unconfirmed-email phrasing. The exact wording can drift across
// Devise versions so we check the conservative-but-stable substring.
func looksLikeUnconfirmed(body []byte) bool {
	lower := strings.ToLower(string(body))
	return strings.Contains(lower, "confirm your email") || strings.Contains(lower, "unconfirmed")
}
