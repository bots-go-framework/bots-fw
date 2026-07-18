package botstoken_test

// Edge-branch tests that push botstoken to 100% statement coverage.
// These complement the existing token_test.go without modifying it.

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw/botstoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encodeRaw base64url-encodes a raw wire string for use as a signed-token input.
func encodeRaw(raw string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// --- parseRaw: len(parts) < 2 branch (line 223-224 in token.go) ---

// TestDecode_onlyVerb_returnsError exercises parseRaw when the raw string has
// no tab separator, so Split yields only one part (just the verb).
func TestDecode_onlyVerb_returnsError(t *testing.T) {
	_, err := botstoken.Decode("justverb")
	require.Error(t, err)
	assert.ErrorIs(t, err, botstoken.ErrInvalidToken)
}

// --- DecodeSignedToken error branches ---

// TestDecodeSignedToken_validBase64ButParseRawFails covers the path where
// base64 decode succeeds but parseRaw fails (e.g. only one field / no tab).
func TestDecodeSignedToken_validBase64ButParseRawFails(t *testing.T) {
	// "noverb" encodes to a valid base64 string, but has no tab → parseRaw fails.
	tok := encodeRaw("noverb")
	_, err := botstoken.DecodeSignedToken(tok, time.Hour, testKey)
	require.Error(t, err)
	assert.ErrorIs(t, err, botstoken.ErrInvalidToken)
}

// TestDecodeSignedToken_missingSig covers the sigHex == "" branch.
// Craft a token that has verb+subject+at but no sig field.
func TestDecodeSignedToken_missingSig(t *testing.T) {
	atVal := strconv.FormatInt(time.Now().Unix(), 10)
	// Wire: "verb\tsubject\tat=<unix>"  — valid structure, has at, missing sig
	raw := fmt.Sprintf("verb\tsubject\tat=%s", atVal)
	tok := encodeRaw(raw)
	_, err := botstoken.DecodeSignedToken(tok, time.Hour, testKey)
	require.Error(t, err)
	assert.ErrorIs(t, err, botstoken.ErrInvalidToken)
}

// TestDecodeSignedToken_missingAt covers the atStr == "" branch.
// Craft a token that has verb+subject+sig but no at field.
func TestDecodeSignedToken_missingAt(t *testing.T) {
	// sig is present but at is absent
	raw := "verb\tsubject\tsig=deadbeef"
	tok := encodeRaw(raw)
	_, err := botstoken.DecodeSignedToken(tok, time.Hour, testKey)
	require.Error(t, err)
	assert.ErrorIs(t, err, botstoken.ErrInvalidToken)
}

// TestDecodeSignedToken_invalidAt covers the strconv.ParseInt error branch.
// Craft a token with sig present and at set to a non-integer.
func TestDecodeSignedToken_invalidAt(t *testing.T) {
	raw := "verb\tsubject\tat=notanumber\tsig=deadbeef"
	tok := encodeRaw(raw)
	_, err := botstoken.DecodeSignedToken(tok, time.Hour, testKey)
	require.Error(t, err)
	assert.ErrorIs(t, err, botstoken.ErrInvalidToken)
}

// TestDecodeSignedToken_malformedSigHex covers the hex.DecodeString error branch.
// The token must pass expiry and key-lookup checks, so we need:
//   - a known key provider
//   - a fresh timestamp (not expired)
//   - a sig value that is non-empty but not valid hex
func TestDecodeSignedToken_malformedSigHex(t *testing.T) {
	atVal := strconv.FormatInt(time.Now().Unix(), 10)
	// "gg" is not valid hex
	raw := fmt.Sprintf("verb\tsubject\tat=%s\tsig=ggzz", atVal)
	tok := encodeRaw(raw)

	// Use a key provider whose VerifyingKey always returns a key (empty keyID)
	kp := &staticKeyProvider{key: []byte("any-key"), keyID: ""}
	_, err := botstoken.DecodeSignedToken(tok, time.Hour, kp)
	require.Error(t, err)
	assert.ErrorIs(t, err, botstoken.ErrInvalidToken)
}

// TestDecodeSignedToken_noArgs_cleanIsNil covers the len(clean)==0→nil branch.
// A successful round-trip with no user args should return Token.Args == nil.
func TestDecodeSignedToken_noArgs_cleanIsNil(t *testing.T) {
	now := time.Now()
	// EncodeSignedToken with nil args; signed tokens only contain internal fields
	// (at, kid, sig). After stripping those, clean should be empty → nil.
	tok, err := botstoken.EncodeSignedToken("ping", "pong", nil, now, testKey)
	require.NoError(t, err)

	decoded, err := botstoken.DecodeSignedToken(tok, time.Hour, testKey)
	require.NoError(t, err)
	assert.Equal(t, "ping", decoded.Verb)
	assert.Equal(t, "pong", decoded.Subject)
	// Args must be nil (not an empty map) when there are no user args.
	assert.Nil(t, decoded.Args, "expected nil Args when token carries no user args")
}

// TestDecodeSignedToken_withoutKeyID ensures that tokens signed without a keyID
// still verify correctly (kid field absent → VerifyingKey called with "").
func TestDecodeSignedToken_withoutKeyID(t *testing.T) {
	kp := &staticKeyProvider{key: []byte("secret"), keyID: ""}
	tok, err := botstoken.EncodeSignedToken("action", "resource", map[string]string{"n": "1"}, time.Now(), kp)
	require.NoError(t, err)

	decoded, err := botstoken.DecodeSignedToken(tok, time.Hour, kp)
	require.NoError(t, err)
	assert.Equal(t, "action", decoded.Verb)
	assert.Equal(t, "1", decoded.Args["n"])
}

// TestDecodeSignedToken_unknownKeyID_returnsInvalidSignature verifies that when
// kp.VerifyingKey returns nil the function returns ErrInvalidSignature.
func TestDecodeSignedToken_unknownKeyID_returnsInvalidSignature(t *testing.T) {
	// Build a valid token with keyID "k1".
	tok, err := botstoken.EncodeSignedToken("do", "s", nil, time.Now(), testKey)
	require.NoError(t, err)

	// A provider that knows "k2" but not "k1" → VerifyingKey returns nil.
	otherKP := &staticKeyProvider{key: []byte("other-secret"), keyID: "k2"}
	_, err = botstoken.DecodeSignedToken(tok, time.Hour, otherKP)
	require.Error(t, err)
	assert.ErrorIs(t, err, botstoken.ErrInvalidSignature)
}

// TestDecodeSignedToken_wrongSig covers the !hmac.Equal branch (ErrInvalidSignature)
// when the signature hex is valid but the HMAC does not match.
func TestDecodeSignedToken_wrongSig(t *testing.T) {
	atVal := strconv.FormatInt(time.Now().Unix(), 10)
	// Build a token with a structurally valid but wrong 32-byte sig (64 hex chars).
	wrongSig := hex.EncodeToString(make([]byte, 32)) // all zeros
	raw := fmt.Sprintf("verb\tsubject\tat=%s\tsig=%s", atVal, wrongSig)
	tok := encodeRaw(raw)

	kp := &staticKeyProvider{key: []byte("any-key"), keyID: ""}
	_, err := botstoken.DecodeSignedToken(tok, time.Hour, kp)
	require.Error(t, err)
	assert.ErrorIs(t, err, botstoken.ErrInvalidSignature)
}
