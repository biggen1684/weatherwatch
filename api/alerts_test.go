package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectNOAA(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		t.Setenv("WEATHERWATCH_USER_AGENT", "weatherwatch (test@example.com)")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"features": [
					{
						"properties": {
							"id": "urn:oid:1",
							"event": "Tornado Warning",
							"geocode": {"UGC": ["FLZ112"]},
							"updated": "2026-06-14T16:08:15+00:00"
						}
					}
				],
				"title": "Current alerts",
				"updated": "2026-06-14T16:08:15+00:00"
			}`))
		}))
		defer server.Close()

		cfg := Config{Area: "FL", Zone: "FLZ112", Events: []string{"Tornado Warning"}}
		alerts, err := ConnectNOAA(server.Client(), server.URL+"/alerts/", cfg, false)
		assert.NoError(t, err)
		assert.Len(t, alerts.Features, 1)
		assert.Equal(t, "Tornado Warning", alerts.Features[0].Properties.Event)
	})

	t.Run("non-200 status", func(t *testing.T) {
		t.Setenv("WEATHERWATCH_USER_AGENT", "weatherwatch (test@example.com)")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := Config{Area: "FL", Zone: "FLZ112", Events: []string{"Tornado Warning"}}
		_, err := ConnectNOAA(server.Client(), server.URL+"/alerts/", cfg, false)
		assert.Error(t, err)
	})

	t.Run("malformed json", func(t *testing.T) {
		t.Setenv("WEATHERWATCH_USER_AGENT", "weatherwatch (test@example.com)")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid`))
		}))
		defer server.Close()

		cfg := Config{Area: "FL", Zone: "FLZ112", Events: []string{"Tornado Warning"}}
		_, err := ConnectNOAA(server.Client(), server.URL+"/alerts/", cfg, false)
		assert.Error(t, err)
	})

	t.Run("missing user agent", func(t *testing.T) {
		t.Setenv("WEATHERWATCH_USER_AGENT", "")

		cfg := Config{Area: "FL", Zone: "FLZ112", Events: []string{"Tornado Warning"}}
		_, err := ConnectNOAA(http.DefaultClient, "https://example.com/alerts/", cfg, false)
		assert.Error(t, err)
	})
}

var mockAlertResponse = AlertResponse{
	Features: []Feature{
		{
			Properties: AlertProperties{
				ID:      "urn:oid:1",
				Event:   "Tornado Warning",
				Geocode: Geocode{UGC: []string{"FLZ108", "FLZ112", "FLZ114"}},
			},
		},
		{
			Properties: AlertProperties{
				ID:      "urn:oid:2",
				Event:   "Heat Advisory",
				Geocode: Geocode{UGC: []string{"FLZ112"}},
			},
		},
		{
			Properties: AlertProperties{
				ID:      "urn:oid:3",
				Event:   "Tornado Warning",
				Geocode: Geocode{UGC: []string{"FLZ069", "FLZ075"}},
			},
		},
	},
}

var mockCfg = Config{
	Zone:   "FLZ112",
	Area:   "FL",
	Events: []string{"Tornado Warning"},
}

func TestFilterAlerts(t *testing.T) {
	t.Run("matches zone and event", func(t *testing.T) {
		matches := FilterAlerts(mockAlertResponse, mockCfg)
		assert.Len(t, matches, 1)
		assert.Equal(t, "Tornado Warning", matches[0].Event)
		assert.Equal(t, "urn:oid:1", matches[0].ID)
	})

	t.Run("wrong zone not matched", func(t *testing.T) {
		cfg := mockCfg
		cfg.Zone = "FLZ999"
		matches := FilterAlerts(mockAlertResponse, cfg)
		assert.Len(t, matches, 0)
	})

	t.Run("wrong event not matched", func(t *testing.T) {
		cfg := mockCfg
		cfg.Events = []string{"Hurricane Warning"}
		matches := FilterAlerts(mockAlertResponse, cfg)
		assert.Len(t, matches, 0)
	})

	t.Run("multiple events in config", func(t *testing.T) {
		cfg := mockCfg
		cfg.Events = []string{"Tornado Warning", "Heat Advisory"}
		matches := FilterAlerts(mockAlertResponse, cfg)
		assert.Len(t, matches, 2)
	})

	t.Run("empty features returns no matches", func(t *testing.T) {
		matches := FilterAlerts(AlertResponse{}, mockCfg)
		assert.Len(t, matches, 0)
	})

	t.Run("correct zone wrong event not matched", func(t *testing.T) {
		cfg := mockCfg
		cfg.Events = []string{"Hurricane Warning"}
		matches := FilterAlerts(mockAlertResponse, cfg)
		assert.Len(t, matches, 0)
	})
}
