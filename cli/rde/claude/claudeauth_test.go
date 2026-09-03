package claude

import (
	"testing"
)

func TestParseClaudeAccessToken(t *testing.T) {
	if got, ok := parseClaudeAccessToken([]byte(`{"claudeAiOauth":{"accessToken":"tok-123","refreshToken":"r"}}`)); !ok || got != "tok-123" {
		t.Errorf("json blob: got %q ok=%v, want tok-123 true", got, ok)
	}
	if got, ok := parseClaudeAccessToken([]byte("  bare-token\n")); !ok || got != "bare-token" {
		t.Errorf("bare token: got %q ok=%v, want bare-token true", got, ok)
	}
	if _, ok := parseClaudeAccessToken([]byte(`{"claudeAiOauth":{"accessToken":""}}`)); ok {
		t.Error("empty accessToken JSON should not be ok")
	}
	if _, ok := parseClaudeAccessToken([]byte("")); ok {
		t.Error("empty input should not be ok")
	}
}

func TestExistingLocalCredentialEnv(t *testing.T) {
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "oauth-tok")
	t.Setenv("ANTHROPIC_API_KEY", "api-key")
	cred, ok := existingLocalCredential()
	if !ok || cred.EnvVar != "CLAUDE_CODE_OAUTH_TOKEN" || cred.Value != "oauth-tok" {
		t.Errorf("oauth env should win: %+v ok=%v", cred, ok)
	}

	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")
	cred, ok = existingLocalCredential()
	if !ok || cred.EnvVar != "ANTHROPIC_API_KEY" || cred.Value != "api-key" {
		t.Errorf("api key fallback: %+v ok=%v", cred, ok)
	}
}

func TestExtractToken(t *testing.T) {
	// Token surrounded by escape sequences, with terminal-restore codes as the
	// trailing output (the case that previously got saved as the "token").
	withEscapes := "\x1b[?2004hPaste here:\x1b[0m\nsk-ant-oat01-abc_DEF-123.xyz\n\x1b[>4m\x1b[<u\x1b[?1004l\x1b[?2031l\x1b[?2004l"
	if got := extractToken(withEscapes); got != "sk-ant-oat01-abc_DEF-123.xyz" {
		t.Errorf("with escapes: got %q", got)
	}
	// Pure escape-sequence garbage must not be mistaken for a token.
	if got := extractToken("\x1b[>4m\x1b[<u\x1b[?1004l\x1b[?2031l\x1b[?2004l"); got != "" {
		t.Errorf("escape-only: got %q, want empty", got)
	}
	// Fallback: a clean token-shaped final line with no sk-ant prefix.
	if got := extractToken("info\nABCDEF0123456789abcdefXYZ\n"); got != "ABCDEF0123456789abcdefXYZ" {
		t.Errorf("fallback: got %q", got)
	}
}
