package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MarkelovSergey/gofermart/internal/auth"
)

func TestAuthMiddleware(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			t.Error("GetUserID() should return true")
		}
		if userID != 123 {
			t.Errorf("userID = %v, want 123", userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtManager)(handler)

	t.Run("valid token in Authorization header", func(t *testing.T) {
		token, _ := jwtManager.GenerateToken(123)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
		}
	})

	t.Run("no token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("malformed Authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "invalid-format")

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusUnauthorized)
		}
	})
}

func TestGetUserID(t *testing.T) {
	t.Run("user ID present", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, int(123))
		userID, ok := GetUserID(ctx)

		if !ok {
			t.Error("GetUserID() ok = false, want true")
		}
		if userID != 123 {
			t.Errorf("userID = %v, want 123", userID)
		}
	})

	t.Run("user ID not present", func(t *testing.T) {
		ctx := context.Background()
		_, ok := GetUserID(ctx)

		if ok {
			t.Error("GetUserID() ok = true, want false")
		}
	})

	t.Run("wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, "not-an-int")
		_, ok := GetUserID(ctx)

		if ok {
			t.Error("GetUserID() ok = true, want false for wrong type")
		}
	})
}

func TestGzipMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	middleware := GzipMiddleware(handler)

	t.Run("without gzip", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
		}

		if rr.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Content-Encoding should not be gzip when not requested")
		}
	})

	t.Run("with Accept-Encoding gzip", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
		}

		if rr.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Content-Encoding should be gzip when requested")
		}
	})
}

func TestLoggingMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := LoggingMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
	}
}
