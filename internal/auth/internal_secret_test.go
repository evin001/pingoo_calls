package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireInternalSecret(t *testing.T) {
	nextCalled := false
	handler := RequireInternalSecret("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))

	t.Run("missing secret is rejected", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d", rec.Code)
		}
		if nextCalled {
			t.Fatal("next handler was called")
		}
	})

	t.Run("bad secret is rejected", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(InternalSecretHeader, "bad")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d", rec.Code)
		}
		if nextCalled {
			t.Fatal("next handler was called")
		}
	})

	t.Run("valid secret passes", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(InternalSecretHeader, "secret")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d", rec.Code)
		}
		if !nextCalled {
			t.Fatal("next handler was not called")
		}
	})
}
