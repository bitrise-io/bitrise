package oauth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
)

// loginTimeout bounds the whole browser round-trip.
const loginTimeout = 5 * time.Minute

// ErrLoginRequired is returned by EnsureFreshPAT when an OAuth-managed token
// can no longer be refreshed (the refresh token is gone or rejected) and the
// user must sign in again.
var ErrLoginRequired = errors.New("OAuth session expired — run 'bitrise auth login --oauth' to sign in again")

// Login runs the browser authorization + token exchange and returns a
// populated auth.Auth (PAT + JWT + refresh token + expiries) without
// persisting it — the caller saves the result. openBrowser opens the
// authorize URL (nil to skip auto-open); the URL is also written to stderr.
func (c Config) Login(ctx context.Context, openBrowser func(string) error, stderr io.Writer) (auth.Auth, error) {
	if c.Issuer == "" {
		return auth.Auth{}, errors.New("OAuth login is not configured: no issuer (set BITRISE_OAUTH_ISSUER)")
	}
	if c.ClientID == "" {
		return auth.Auth{}, errors.New("OAuth login is not available in this build: no client_id is compiled in. Use 'auth login' or 'auth login --email' instead")
	}

	state, err := newState()
	if err != nil {
		return auth.Auth{}, err
	}
	verifier, challenge, err := newPKCE()
	if err != nil {
		return auth.Auth{}, err
	}

	cs, err := newCallbackServer(state)
	if err != nil {
		return auth.Auth{}, err
	}
	defer cs.close()
	cs.start()

	authURL := c.authorizeURL(challenge, state, cs.redirectURI())
	if _, err := fmt.Fprintf(stderr, "Opening your browser to sign in to Bitrise.\nIf it doesn't open automatically, visit:\n\n  %s\n\n", authURL); err != nil {
		return auth.Auth{}, err
	}
	if openBrowser != nil {
		if err := openBrowser(authURL); err != nil {
			if _, werr := fmt.Fprintf(stderr, "(couldn't open the browser automatically: %v)\n", err); werr != nil {
				return auth.Auth{}, werr
			}
		}
	}

	waitCtx, cancel := context.WithTimeout(ctx, loginTimeout)
	defer cancel()
	code, err := cs.wait(waitCtx)
	if err != nil {
		return auth.Auth{}, err
	}

	jwtResp, err := c.exchangeCodeForJWT(ctx, code, verifier, cs.redirectURI())
	if err != nil {
		return auth.Auth{}, fmt.Errorf("exchange authorization code: %w", err)
	}
	pat, patExpiry, err := c.exchangeJWTForPAT(ctx, jwtResp.AccessToken)
	if err != nil {
		return auth.Auth{}, fmt.Errorf("exchange token for a Bitrise PAT: %w", err)
	}

	now := time.Now()
	return auth.Auth{
		Token:        pat,
		TokenExpiry:  patExpiry,
		JWT:          jwtResp.AccessToken,
		JWTExpiry:    jwtExpiry(jwtResp, now),
		RefreshToken: jwtResp.RefreshToken,
	}, nil
}

// authorizeURL builds the WorkOS authorize URL; offline_access requests a
// refresh token.
func (c Config) authorizeURL(challenge, state, redirectURI string) string {
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {c.ClientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {"openid offline_access"},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}
	if c.Resource != "" {
		q.Set("resource", c.Resource)
	}
	return c.authorizeEndpoint() + "?" + q.Encode()
}

// EnsureFreshPAT returns a usable PAT, refreshing it without a browser when
// needed. A manually pasted / email-login token (no refresh token) is
// returned untouched. The ladder:
//
//	PAT valid              → return it
//	PAT expired            → exchange JWT → new PAT
//	PAT + JWT expired      → refresh-token grant → new JWT → new PAT
//	refresh token rejected → ErrLoginRequired
func (c Config) EnsureFreshPAT(ctx context.Context, resolvedToken string) (string, error) {
	a, err := auth.Load()
	if err != nil {
		return "", err
	}
	if !a.IsOAuthManaged() {
		// Manual token (paste/email login): use as-is, never refreshed.
		return resolvedToken, nil
	}

	now := time.Now()
	if a.Token != "" && now.Add(refreshSkew).Before(a.TokenExpiry) {
		return a.Token, nil
	}

	// PAT stale. If the JWT is still good, a single exchange refreshes the PAT.
	if a.JWT != "" && now.Add(refreshSkew).Before(a.JWTExpiry) {
		pat, expiry, err := c.exchangeJWTForPAT(ctx, a.JWT)
		if err == nil {
			a.Token, a.TokenExpiry = pat, expiry
			if err := auth.Save(a); err != nil {
				return "", err
			}
			return pat, nil
		}
		// Exchange failed despite an unexpired JWT — fall through and try a
		// full refresh before giving up.
	}

	// PAT and JWT both stale: refresh the JWT, then exchange it.
	if a.RefreshToken == "" {
		return "", ErrLoginRequired
	}
	refreshed, err := c.refreshJWT(ctx, a.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("%w (refresh failed: %v)", ErrLoginRequired, err)
	}
	a.JWT = refreshed.AccessToken
	a.JWTExpiry = jwtExpiry(refreshed, now)
	if refreshed.RefreshToken != "" { // WorkOS may rotate the refresh token
		a.RefreshToken = refreshed.RefreshToken
	}

	pat, expiry, err := c.exchangeJWTForPAT(ctx, a.JWT)
	if err != nil {
		return "", fmt.Errorf("exchange refreshed token for a PAT: %w", err)
	}
	a.Token, a.TokenExpiry = pat, expiry
	if err := auth.Save(a); err != nil {
		return "", err
	}
	return pat, nil
}
