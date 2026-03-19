package botsfw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPingHandler(t *testing.T) {
	t.Run("returns_pong", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		PingHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		body := w.Body.String()
		if body != "Pong" {
			t.Errorf("Expected body 'Pong', got %q", body)
		}
		if cors := resp.Header.Get("Access-Control-Allow-Origin"); cors != "*" {
			t.Errorf("Expected CORS header '*', got %q", cors)
		}
	})
}

func TestNotFoundHandler(t *testing.T) {
	t.Run("returns_404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		w := httptest.NewRecorder()
		NotFoundHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})
}
