package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// tokenResponse covers both the WorkOS token endpoint (code→JWT, refresh) and
// the monolith OIDC exchange (JWT→PAT); fields absent from a response stay zero.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// exchangeCodeForJWT trades a code for a JWT + refresh token. redirectURI
// must match the one sent on the authorize request.
func (c Config) exchangeCodeForJWT(ctx context.Context, code, verifier, redirectURI string) (tokenResponse, error) {
	return c.postForm(ctx, c.tokenEndpoint(), url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {c.ClientID},
		"code_verifier": {verifier},
	})
}

// refreshJWT obtains a fresh JWT (and possibly a rotated refresh token).
func (c Config) refreshJWT(ctx context.Context, refreshToken string) (tokenResponse, error) {
	return c.postForm(ctx, c.tokenEndpoint(), url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {c.ClientID},
	})
}

// exchangeJWTForPAT trades a JWT for a Bitrise PAT (RFC 8693), mirroring the
// MCP server's callExchangeEndpoint: form-encoded, no client_id/resource
// (the audience rides inside the JWT). Returns the PAT and its expiry.
func (c Config) exchangeJWTForPAT(ctx context.Context, jwt string) (string, time.Time, error) {
	resp, err := c.postForm(ctx, c.OIDCTokenEndpoint, url.Values{
		"grant_type":         {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"subject_token":      {jwt},
		"subject_token_type": {"urn:ietf:params:oauth:token-type:access_token"},
	})
	if err != nil {
		return "", time.Time{}, err
	}
	if resp.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("OIDC exchange response missing access_token")
	}
	expiry := time.Now().Add(defaultPATLifetime)
	if resp.ExpiresIn > 0 {
		expiry = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	}
	return resp.AccessToken, expiry, nil
}

func (c Config) postForm(ctx context.Context, endpoint string, form url.Values) (tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return tokenResponse{}, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("read token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return tokenResponse{}, fmt.Errorf("token endpoint %s returned %d: %s", endpoint, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return tokenResponse{}, fmt.Errorf("parse token response: %w", err)
	}
	return tr, nil
}

// jwtExpiry prefers expires_in, falls back to the JWT's exp claim, and
// finally to a short conservative window so a refresh is attempted soon
// rather than never.
func jwtExpiry(resp tokenResponse, now time.Time) time.Time {
	if resp.ExpiresIn > 0 {
		return now.Add(time.Duration(resp.ExpiresIn) * time.Second)
	}
	if exp, ok := parseJWTExp(resp.AccessToken); ok {
		return exp
	}
	return now.Add(5 * time.Minute)
}

// parseJWTExp decodes the exp claim from a JWT without verifying its signature
// (the monolith verifies it; the CLI only needs the expiry for scheduling).
func parseJWTExp(token string) (time.Time, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, false
	}
	payload := parts[1]
	if pad := len(payload) % 4; pad != 0 {
		payload += strings.Repeat("=", 4-pad)
	}
	data, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return time.Time{}, false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(data, &claims); err != nil || claims.Exp == 0 {
		return time.Time{}, false
	}
	return time.Unix(claims.Exp, 0), true
}
