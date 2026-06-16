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

var mockCfg = Config{
	Zone:   "FLZ112",
	Area:   "FL",
	Events: []string{"Tornado Warning"},
}

var mockFlatAlerts = []FlatAlert{
	{
		ID:    "urn:oid:1",
		Event: "Tornado Warning",
		Zones: []string{"FLZ108", "FLZ112", "FLZ114"},
	},
	{
		ID:    "urn:oid:2",
		Event: "Heat Advisory",
		Zones: []string{"FLZ112"},
	},
	{
		ID:    "urn:oid:3",
		Event: "Tornado Warning",
		Zones: []string{"FLZ069", "FLZ075"},
	},
}

func TestFilterAlerts(t *testing.T) {
	t.Run("matches zone and event", func(t *testing.T) {
		matches := FilterAlerts(mockFlatAlerts, mockCfg)
		assert.Len(t, matches, 1)
		assert.Equal(t, "Tornado Warning", matches[0].Event)
		assert.Equal(t, "urn:oid:1", matches[0].ID)
	})

	t.Run("wrong zone not matched", func(t *testing.T) {
		cfg := mockCfg
		cfg.Zone = "FLZ999"
		matches := FilterAlerts(mockFlatAlerts, cfg)
		assert.Len(t, matches, 0)
	})

	t.Run("wrong event not matched", func(t *testing.T) {
		cfg := mockCfg
		cfg.Events = []string{"Hurricane Warning"}
		matches := FilterAlerts(mockFlatAlerts, cfg)
		assert.Len(t, matches, 0)
	})

	t.Run("multiple events in config", func(t *testing.T) {
		cfg := mockCfg
		cfg.Events = []string{"Tornado Warning", "Heat Advisory"}
		matches := FilterAlerts(mockFlatAlerts, cfg)
		assert.Len(t, matches, 2)
	})

	t.Run("empty alerts returns no matches", func(t *testing.T) {
		matches := FilterAlerts([]FlatAlert{}, mockCfg)
		assert.Len(t, matches, 0)
	})

	t.Run("correct zone wrong event not matched", func(t *testing.T) {
		cfg := mockCfg
		cfg.Events = []string{"Hurricane Warning"}
		matches := FilterAlerts(mockFlatAlerts, cfg)
		assert.Len(t, matches, 0)
	})
}

func TestFlattenAlerts(t *testing.T) {
	t.Run("flattens nested response into flat alerts", func(t *testing.T) {
		resp := AlertResponse{
			Features: []Feature{
				{
					Properties: AlertProperties{
						ID:      "urn:oid:1",
						Event:   "Tornado Warning",
						Geocode: Geocode{UGC: []string{"FLZ108", "FLZ112"}},
					},
				},
				{
					Properties: AlertProperties{
						ID:      "urn:oid:2",
						Event:   "Heat Advisory",
						Geocode: Geocode{UGC: []string{"FLZ112"}},
					},
				},
			},
		}

		flat := FlattenAlerts(resp)

		assert.Len(t, flat, 2)
		assert.Equal(t, "urn:oid:1", flat[0].ID)
		assert.Equal(t, "Tornado Warning", flat[0].Event)
		assert.Equal(t, []string{"FLZ108", "FLZ112"}, flat[0].Zones)
		assert.Equal(t, "urn:oid:2", flat[1].ID)
		assert.Equal(t, "Heat Advisory", flat[1].Event)
	})

	t.Run("empty features returns empty slice", func(t *testing.T) {
		flat := FlattenAlerts(AlertResponse{})
		assert.Len(t, flat, 0)
	})
}
