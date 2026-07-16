// Package botstoken implements a compact, URL-safe token codec for
// (verb string, subject string, args map[string]string) triples.
//
// Two token variants are provided:
//
//  1. Plain tokens — encode verb/subject/args into a short string, optionally
//     base64url-encoded. Suitable for in-platform use (Telegram callback_data,
//     WhatsApp interactive reply id) where the platform protects integrity.
//     Guaranteed ≤ 64 bytes for typical inputs; Encode returns an error if the
//     result would exceed 64 bytes.
//
//  2. Signed tokens — HMAC-SHA256 signed with an issued-at timestamp.
//     Suitable for tokens that leave the platform (wa.me URLs, web deep links).
//     Expiry is checked on Decode. The key is provided via a pluggable KeyProvider.
//
// Token wire format (plain, before base64url):
//
//	<verb>\t<subject>[\t<k1>=<v1>\t<k2>=<v2>...]
//
// Signed token wire format (base64url, no padding):
//
//	b64url(<verb>\t<subject>\t<k>=<v>...\tat=<unix-seconds>\tsig=<hex-hmac-sha256>)
//
// Args are sorted by key for deterministic encoding.
package botstoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// MaxTokenBytes is the maximum allowed byte length of the encoded token.
	MaxTokenBytes = 64

	fieldSep = '\t'
	kvSep    = '='
)

// ErrTokenTooLong is returned when an encoded token would exceed MaxTokenBytes.
var ErrTokenTooLong = errors.New("botstoken: encoded token exceeds 64 bytes")

// ErrInvalidToken is returned when a token cannot be parsed.
var ErrInvalidToken = errors.New("botstoken: invalid token format")

// ErrTokenExpired is returned when the issued-at timestamp is too old.
var ErrTokenExpired = errors.New("botstoken: token has expired")

// ErrInvalidSignature is returned when the HMAC signature does not match.
var ErrInvalidSignature = errors.New("botstoken: invalid signature")

// Token holds the decoded fields of a token.
type Token struct {
	Verb    string
	Subject string
	Args    map[string]string
}

// -------------------------------------------------------------------
// Plain token codec
// -------------------------------------------------------------------

// Encode encodes (verb, subject, args) into a compact string token.
// Args are sorted by key for deterministic output.
// Returns ErrTokenTooLong if the result would exceed 64 bytes.
func Encode(verb, subject string, args map[string]string) (string, error) {
	raw := buildRaw(verb, subject, args)
	if len(raw) > MaxTokenBytes {
		return "", fmt.Errorf("%w: got %d bytes (verb=%q subject=%q)", ErrTokenTooLong, len(raw), verb, subject)
	}
	return raw, nil
}

// Decode parses a plain token produced by Encode.
func Decode(token string) (Token, error) {
	return parseRaw(token)
}

// -------------------------------------------------------------------
// Signed token codec
// -------------------------------------------------------------------

// KeyProvider returns the HMAC signing key.
// Implementations may rotate keys; they receive the key ID stored in the token.
// For the current (signing) call, keyID is empty.
type KeyProvider interface {
	// SigningKey returns the current signing key and its ID.
	SigningKey() (key []byte, keyID string)
	// VerifyingKey returns the key for the given key ID.
	// Return nil to indicate that the key ID is unknown.
	VerifyingKey(keyID string) []byte
}

// EncodeSignedToken encodes and HMAC-signs a token with an issued-at timestamp.
// The result is base64url-encoded (no padding). Signed tokens carry authentication
// overhead (HMAC-SHA256 signature + issued-at timestamp) that makes them inherently
// larger than plain tokens; they are intended for off-platform use (web, wa.me deep
// links) where URL length is not as constrained as in-platform callback fields.
// There is no explicit length limit on signed tokens — the caller is responsible
// for keeping verb, subject and args short enough for the target channel.
func EncodeSignedToken(verb, subject string, args map[string]string, now time.Time, kp KeyProvider) (string, error) {
	key, keyID := kp.SigningKey()
	argsWithMeta := copyArgs(args)
	argsWithMeta["at"] = strconv.FormatInt(now.Unix(), 10)
	if keyID != "" {
		argsWithMeta["kid"] = keyID
	}

	payload := buildRaw(verb, subject, argsWithMeta)
	sig := computeHMAC(key, payload)
	argsWithMeta["sig"] = hex.EncodeToString(sig)

	raw := buildRaw(verb, subject, argsWithMeta)
	encoded := base64.RawURLEncoding.EncodeToString([]byte(raw))
	return encoded, nil
}

// DecodeSignedToken decodes and verifies a signed token produced by EncodeSignedToken.
// Returns ErrTokenExpired if the token is older than maxAge.
// Returns ErrInvalidSignature if the HMAC is wrong.
// Returns ErrInvalidToken if the format is unrecognised.
func DecodeSignedToken(token string, maxAge time.Duration, kp KeyProvider) (Token, error) {
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Token{}, fmt.Errorf("%w: base64 decode failed: %v", ErrInvalidToken, err)
	}
	t, err := parseRaw(string(b))
	if err != nil {
		return Token{}, err
	}

	// Extract and remove "sig" and "at" (and optionally "kid") before verification.
	sigHex := t.Args["sig"]
	atStr := t.Args["at"]
	keyID := t.Args["kid"]

	if sigHex == "" {
		return Token{}, fmt.Errorf("%w: missing sig field", ErrInvalidToken)
	}
	if atStr == "" {
		return Token{}, fmt.Errorf("%w: missing at field", ErrInvalidToken)
	}

	atUnix, convErr := strconv.ParseInt(atStr, 10, 64)
	if convErr != nil {
		return Token{}, fmt.Errorf("%w: invalid at field: %v", ErrInvalidToken, convErr)
	}

	// Verify expiry.
	issuedAt := time.Unix(atUnix, 0)
	if time.Since(issuedAt) > maxAge {
		return Token{}, ErrTokenExpired
	}

	// Reconstruct payload as it was before sig was added.
	argsForVerify := copyArgs(t.Args)
	delete(argsForVerify, "sig")
	payload := buildRaw(t.Verb, t.Subject, argsForVerify)

	// Verify signature.
	key := kp.VerifyingKey(keyID)
	if key == nil {
		return Token{}, fmt.Errorf("%w: unknown key id %q", ErrInvalidSignature, keyID)
	}
	expected := computeHMAC(key, payload)
	gotSig, hexErr := hex.DecodeString(sigHex)
	if hexErr != nil {
		return Token{}, fmt.Errorf("%w: malformed sig hex: %v", ErrInvalidToken, hexErr)
	}
	if !hmac.Equal(expected, gotSig) {
		return Token{}, ErrInvalidSignature
	}

	// Return clean token without internal fields.
	clean := copyArgs(t.Args)
	delete(clean, "sig")
	delete(clean, "at")
	delete(clean, "kid")
	if len(clean) == 0 {
		clean = nil
	}
	return Token{Verb: t.Verb, Subject: t.Subject, Args: clean}, nil
}

// -------------------------------------------------------------------
// Internal helpers
// -------------------------------------------------------------------

// buildRaw assembles the raw (wire) token string from its components.
// All fields are sorted for deterministic output.
func buildRaw(verb, subject string, args map[string]string) string {
	var sb strings.Builder
	sb.WriteString(verb)
	sb.WriteByte(fieldSep)
	sb.WriteString(subject)

	if len(args) > 0 {
		keys := sortedKeys(args)
		for _, k := range keys {
			sb.WriteByte(fieldSep)
			sb.WriteString(k)
			sb.WriteByte(kvSep)
			sb.WriteString(args[k])
		}
	}
	return sb.String()
}

// parseRaw parses a raw token string.
func parseRaw(raw string) (Token, error) {
	if raw == "" {
		return Token{}, ErrInvalidToken
	}
	parts := strings.Split(raw, string(fieldSep))
	if len(parts) < 2 {
		return Token{}, fmt.Errorf("%w: expected at least verb and subject", ErrInvalidToken)
	}
	t := Token{
		Verb:    parts[0],
		Subject: parts[1],
	}
	if len(parts) > 2 {
		t.Args = make(map[string]string, len(parts)-2)
		for _, kv := range parts[2:] {
			idx := strings.IndexByte(kv, kvSep)
			if idx < 0 {
				return Token{}, fmt.Errorf("%w: malformed arg %q (missing '=')", ErrInvalidToken, kv)
			}
			t.Args[kv[:idx]] = kv[idx+1:]
		}
	}
	return t, nil
}

func computeHMAC(key []byte, payload string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return mac.Sum(nil)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func copyArgs(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
