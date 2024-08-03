package server_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kkstas/tjener/internal/server"
)

func TestHome(t *testing.T) {
	t.Run("responds with html", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/home", nil)
		server.NewApplication(nil).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)

		contentType := response.Header().Get("content-type")

		if !strings.HasPrefix(contentType, "text/html") {
			t.Errorf("invalid content type %q", contentType)

		}
	})
}

func TestHealthCheck(t *testing.T) {
	t.Run("returns status 200", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/health-check", nil)
		server.NewApplication(nil).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
	})
}

func TestStaticCss(t *testing.T) {
	t.Run("returns css file content with status 200", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/static/css/out", nil)
		server.NewApplication(nil).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)

		if response.Body.Len() == 0 {
			t.Errorf("response body is empty")
		}
	})
}

func TestNotFound(t *testing.T) {
	s := server.NewApplication(nil)

	cases := []struct {
		method string
		target string
	}{
		{http.MethodGet, "/"},
		{http.MethodPost, "/"},
		{http.MethodPut, "/"},
		{http.MethodPatch, "/"},
		{http.MethodDelete, "/"},
		{http.MethodGet, "/abcd1234"},
		{http.MethodPost, "/abcd1234"},
		{http.MethodPut, "/abcd1234"},
		{http.MethodPatch, "/abcd1234"},
		{http.MethodDelete, "/abcd1234"},
	}

	for _, want := range cases {
		t.Run(fmt.Sprintf("responds with 404 for '%s %s'", want.method, want.target), func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(want.method, want.target, nil)
			s.ServeHTTP(response, request)
			assertStatus(t, response.Code, http.StatusNotFound)
		})
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}
