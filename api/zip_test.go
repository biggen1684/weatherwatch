package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateZip(t *testing.T) {
	t.Run("valid zip", func(t *testing.T) {
		err := validateZip("32547")
		assert.NoError(t, err)
	})

	t.Run("too short", func(t *testing.T) {
		err := validateZip("1234")
		assert.Error(t, err)
	})

	t.Run("too long", func(t *testing.T) {
		err := validateZip("123456")
		assert.Error(t, err)
	})

	t.Run("contains letters", func(t *testing.T) {
		err := validateZip("1234a")
		assert.Error(t, err)
	})

	t.Run("contains symbols", func(t *testing.T) {
		err := validateZip("123-4")
		assert.Error(t, err)
	})
}

func TestZipToLatLon(t *testing.T) {
	t.Run("valid zip returns lat/lon", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"places": [
					{"latitude": "34.1025", "longitude": "-118.4150"}
				]
			}`))
		}))
		defer server.Close()

		lat, lon, err := zipToLatLon(server.Client(), server.URL+"/us/", "90210", false)
		assert.NoError(t, err)
		assert.Equal(t, "34.1025", lat)
		assert.Equal(t, "-118.4150", lon)
	})

	t.Run("zip not found returns 404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		_, _, err := zipToLatLon(server.Client(), server.URL+"/us/", "99999", false)
		assert.Error(t, err)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`internal error`))
		}))
		defer server.Close()

		_, _, err := zipToLatLon(server.Client(), server.URL+"/us/", "90210", false)
		assert.Error(t, err)
	})

	t.Run("malformed json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid`))
		}))
		defer server.Close()

		_, _, err := zipToLatLon(server.Client(), server.URL+"/us/", "90210", false)
		assert.Error(t, err)
	})

	t.Run("empty places array", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"places": []}`))
		}))
		defer server.Close()

		_, _, err := zipToLatLon(server.Client(), server.URL+"/us/", "90210", false)
		assert.Error(t, err)
	})
}
