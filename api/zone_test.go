package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatLonToZone(t *testing.T) {
	t.Run("valid lat/lon returns zone and county", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"properties": {
					"forecastZone": "https://api.weather.gov/zones/forecast/FLZ112",
					"county": "https://api.weather.gov/zones/county/FLC005",
					"type": "land"
				}
			}`))
		}))
		defer server.Close()

		zone, county, err := latLonToZone(server.Client(), server.URL+"/", "test-agent", "30.4593", "-86.6066", false)
		assert.NoError(t, err)
		assert.Equal(t, "FLZ112", zone)
		assert.Equal(t, "FLC005", county)
	})

	t.Run("marine zone returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"properties": {
					"forecastZone": "https://api.weather.gov/zones/forecast/GMZ735",
					"county": "",
					"type": "marine"
				}
			}`))
		}))
		defer server.Close()

		_, _, err := latLonToZone(server.Client(), server.URL+"/", "test-agent", "30.2007", "-85.8136", false)
		assert.Error(t, err)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`internal error`))
		}))
		defer server.Close()

		_, _, err := latLonToZone(server.Client(), server.URL+"/", "test-agent", "30.4593", "-86.6066", false)
		assert.Error(t, err)
	})

	t.Run("malformed json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid`))
		}))
		defer server.Close()

		_, _, err := latLonToZone(server.Client(), server.URL+"/", "test-agent", "30.4593", "-86.6066", false)
		assert.Error(t, err)
	})
}
