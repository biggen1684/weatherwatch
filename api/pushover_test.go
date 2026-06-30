package weather

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendPushover(t *testing.T) {
	t.Run("successful notification", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)

			err := r.ParseForm()
			assert.NoError(t, err)
			assert.Equal(t, "test-api-key", r.FormValue("token"))
			assert.Equal(t, "test-user-key", r.FormValue("user"))
			assert.Equal(t, "[Home] Tornado Warning issued...", r.FormValue("title"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":1,"request":"abc123"}`))
		}))
		defer server.Close()

		alert := AlertProperties{
			Headline:    "Tornado Warning issued...",
			Description: "Dangerous tornado situation.",
			AreaDesc:    "South Walton; Coastal Bay",
		}

		err := SendPushover(server.Client(), server.URL, "test-api-key", "test-user-key", "Home", alert)
		assert.NoError(t, err)
	})

	t.Run("message is truncated if over 1024 chars", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			assert.NoError(t, err)
			message := r.FormValue("message")
			assert.LessOrEqual(t, len(message), 1024)
			assert.True(t, strings.HasSuffix(message, "..."))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":1}`))
		}))
		defer server.Close()

		longDescription := strings.Repeat("A", 1100)
		alert := AlertProperties{
			Headline:    "Test",
			Description: longDescription,
			AreaDesc:    "Test Area",
		}

		err := SendPushover(server.Client(), server.URL, "test-api-key", "test-user-key", "Home", alert)
		assert.NoError(t, err)
	})

	t.Run("invalid credentials returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"user":"invalid","errors":["user identifier is invalid"],"status":0}`))
		}))
		defer server.Close()

		alert := AlertProperties{Headline: "Test", Description: "test"}

		err := SendPushover(server.Client(), server.URL, "test-api-key", "bad-user-key", "Home", alert)
		assert.Error(t, err)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		alert := AlertProperties{Headline: "Test", Description: "test"}

		err := SendPushover(server.Client(), server.URL, "test-api-key", "test-user-key", "Home", alert)
		assert.Error(t, err)
	})
}
