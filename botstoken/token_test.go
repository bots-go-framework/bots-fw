package botstoken_test

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw/botstoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

type staticKeyProvider struct {
	key   []byte
	keyID string
}

func (p *staticKeyProvider) SigningKey() ([]byte, string) { return p.key, p.keyID }
func (p *staticKeyProvider) VerifyingKey(id string) []byte {
	if id == p.keyID {
		return p.key
	}
	return nil
}

var testKey = &staticKeyProvider{key: []byte("test-secret-key"), keyID: "k1"}

// --- Encode / Decode ---

func TestEncode_simple(t *testing.T) {
	tok, err := botstoken.Encode("confirm", "spot-123", nil)
	require.NoError(t, err)
	assert.Equal(t, "confirm\tspot-123", tok)
}

func TestEncode_withArgs(t *testing.T) {
	args := map[string]string{"action": "join", "date": "2026-07-16"}
	tok, err := botstoken.Encode("go", "evt-1", args)
	require.NoError(t, err)
	// args must be deterministically ordered (sorted keys)
	assert.Contains(t, tok, "action=join")
	assert.Contains(t, tok, "date=2026-07-16")
	// second call must produce identical output (deterministic)
	tok2, _ := botstoken.Encode("go", "evt-1", args)
	assert.Equal(t, tok, tok2)
}

func TestEncode_exact64Bytes(t *testing.T) {
	// Craft an input whose raw encoding is exactly 64 bytes.
	// "v\ts" = 3 bytes, so the remaining 61 bytes fill subject.
	verb := "v"
	subject := strings.Repeat("x", 62)
	raw := verb + "\t" + subject // 1 + 1 + 62 = 64 bytes
	assert.Equal(t, 64, len(raw))

	tok, err := botstoken.Encode(verb, subject, nil)
	require.NoError(t, err)
	assert.Len(t, tok, 64)
}

func TestEncode_over64Bytes_returnsError(t *testing.T) {
	// Subject of 63 bytes + "v\t" = 65 bytes → should fail.
	_, err := botstoken.Encode("v", strings.Repeat("x", 63), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, botstoken.ErrTokenTooLong)
}

func TestDecode_simple(t *testing.T) {
	tok, err := botstoken.Encode("cancel", "intent-42", nil)
	require.NoError(t, err)

	t2, err := botstoken.Decode(tok)
	require.NoError(t, err)
	assert.Equal(t, "cancel", t2.Verb)
	assert.Equal(t, "intent-42", t2.Subject)
	assert.Empty(t, t2.Args)
}

func TestDecode_withArgs(t *testing.T) {
	args := map[string]string{"k": "v", "n": "1"}
	tok, err := botstoken.Encode("do", "s1", args)
	require.NoError(t, err)

	t2, err := botstoken.Decode(tok)
	require.NoError(t, err)
	assert.Equal(t, "do", t2.Verb)
	assert.Equal(t, "s1", t2.Subject)
	assert.Equal(t, args, t2.Args)
}

func TestDecode_emptyToken_returnsError(t *testing.T) {
	_, err := botstoken.Decode("")
	assert.ErrorIs(t, err, botstoken.ErrInvalidToken)
}

func TestDecode_malformedArg_returnsError(t *testing.T) {
	// Manually craft a token with a bad arg (no '=').
	_, err := botstoken.Decode("verb\tsubj\tbadarg")
	assert.ErrorIs(t, err, botstoken.ErrInvalidToken)
}

// --- Signed tokens ---

func TestEncodeSignedToken_decodeSuccess(t *testing.T) {
	now := time.Now()
	args := map[string]string{"action": "verify"}
	tok, err := botstoken.EncodeSignedToken("link", "user-99", args, now, testKey)
	require.NoError(t, err)
	require.NotEmpty(t, tok)
	// Signed tokens carry authentication overhead (HMAC + timestamp) so they are
	// intentionally longer than the 64-byte plain-codec limit.

	decoded, err := botstoken.DecodeSignedToken(tok, time.Hour, testKey)
	require.NoError(t, err)
	assert.Equal(t, "link", decoded.Verb)
	assert.Equal(t, "user-99", decoded.Subject)
	assert.Equal(t, "verify", decoded.Args["action"])
	// Internal fields must not leak into the decoded result.
	assert.Empty(t, decoded.Args["sig"])
	assert.Empty(t, decoded.Args["at"])
	assert.Empty(t, decoded.Args["kid"])
}

func TestEncodeSignedToken_deterministic(t *testing.T) {
	now := time.Unix(1700000000, 0)
	args := map[string]string{"x": "1"}
	tok1, err1 := botstoken.EncodeSignedToken("a", "b", args, now, testKey)
	tok2, err2 := botstoken.EncodeSignedToken("a", "b", args, now, testKey)
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, tok1, tok2, "same inputs must produce same token")
}

func TestDecodeSignedToken_expired(t *testing.T) {
	// Issue the token 2 hours in the past.
	past := time.Now().Add(-2 * time.Hour)
	tok, err := botstoken.EncodeSignedToken("x", "y", nil, past, testKey)
	require.NoError(t, err)

	_, err = botstoken.DecodeSignedToken(tok, time.Hour, testKey)
	assert.ErrorIs(t, err, botstoken.ErrTokenExpired)
}

func TestDecodeSignedToken_tampered(t *testing.T) {
	tok, err := botstoken.EncodeSignedToken("do", "obj-1", nil, time.Now(), testKey)
	require.NoError(t, err)

	// Decode the token, modify the verb, re-encode without re-signing.
	// This reliably produces a token whose signature check will fail.
	rawBytes, decErr := base64.RawURLEncoding.DecodeString(tok)
	require.NoError(t, decErr)
	// Replace "do" with "xx" in the raw payload.
	tamperRaw := strings.Replace(string(rawBytes), "do\t", "xx\t", 1)
	tampered := base64.RawURLEncoding.EncodeToString([]byte(tamperRaw))

	_, err = botstoken.DecodeSignedToken(tampered, time.Hour, testKey)
	assert.Error(t, err)
}

func TestDecodeSignedToken_unknownKey(t *testing.T) {
	tok, err := botstoken.EncodeSignedToken("do", "obj-1", nil, time.Now(), testKey)
	require.NoError(t, err)

	noKeyProvider := &staticKeyProvider{key: []byte("other"), keyID: "k2"}
	_, err = botstoken.DecodeSignedToken(tok, time.Hour, noKeyProvider)
	assert.Error(t, err)
}

func TestDecodeSignedToken_invalidBase64(t *testing.T) {
	_, err := botstoken.DecodeSignedToken("!!!not-base64!!!", time.Hour, testKey)
	assert.ErrorIs(t, err, botstoken.ErrInvalidToken)
}

// TestEncodeSignedToken_hasNoLengthLimit verifies that the signed variant does not
// enforce the 64-byte limit (authentication overhead makes it inherently longer).
func TestEncodeSignedToken_hasNoLengthLimit(t *testing.T) {
	// A long subject would exceed 64 bytes in the signed variant, but no error is returned.
	tok, err := botstoken.EncodeSignedToken("v", strings.Repeat("x", 50), nil, time.Now(), testKey)
	assert.NoError(t, err)
	assert.Greater(t, len(tok), botstoken.MaxTokenBytes, "signed token should be > 64 bytes for non-trivial inputs")
}

// TestEncode_noArgsNilAndEmpty checks that nil and empty args produce the same token.
func TestEncode_noArgsNilAndEmpty(t *testing.T) {
	tok1, _ := botstoken.Encode("do", "s", nil)
	tok2, _ := botstoken.Encode("do", "s", map[string]string{})
	assert.Equal(t, tok1, tok2)
}
