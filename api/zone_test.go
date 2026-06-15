package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatLonToZone(t *testing.T) {
	t.Run("valid lat/lon returns zone", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"properties": {
					"forecastZone": "https://api.weather.gov/zones/forecast/CAZ368"
				}
			}`))
		}))
		defer server.Close()

		zone, err := latLonToZone(server.Client(), server.URL+"/", "test-agent", "34.1025", "-118.4150", false)
		assert.NoError(t, err)
		assert.Equal(t, "CAZ368", zone)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`internal error`))
		}))
		defer server.Close()

		_, err := latLonToZone(server.Client(), server.URL+"/", "test-agent", "34.1025", "-118.4150", false)
		assert.Error(t, err)
	})

	t.Run("malformed json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid`))
		}))
		defer server.Close()

		_, err := latLonToZone(server.Client(), server.URL+"/", "test-agent", "30.4593", "-86.6066", false)
		assert.Error(t, err)
	})
}
