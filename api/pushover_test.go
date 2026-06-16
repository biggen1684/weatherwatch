package weather

import (
	"net/http"
	"net/http/httptest"
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
			assert.Equal(t, "Tornado Warning", r.FormValue("title"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":1,"request":"abc123"}`))
		}))
		defer server.Close()

		alert := FlatAlert{
			Event:    "Tornado Warning",
			Headline: "Tornado Warning issued June 14 at 3:26AM EDT...",
		}

		err := SendPushover(server.Client(), server.URL, "test-api-key", "test-user-key", alert)
		assert.NoError(t, err)
	})

	t.Run("invalid user key returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"user":"invalid","errors":["user identifier is invalid"],"status":0}`))
		}))
		defer server.Close()

		alert := FlatAlert{Event: "Tornado Warning", Headline: "test"}

		err := SendPushover(server.Client(), server.URL, "test-api-key", "test-user-key", alert)
		assert.Error(t, err)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		alert := FlatAlert{Event: "Tornado Warning", Headline: "test"}

		err := SendPushover(server.Client(), server.URL, "test-api-key", "test-user-key", alert)
		assert.Error(t, err)
	})
}
