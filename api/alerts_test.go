package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var mockLocation = Location{
	Name:   "Home",
	Area:   "FL",
	Zone:   "FLZ112",
	County: "FLC005",
}

var mockEvents = []string{"Tornado Warning"}

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

func TestFilterAlerts(t *testing.T) {
	t.Run("matches zone and event", func(t *testing.T) {
		matches := FilterAlerts(mockAlertResponse, mockLocation, mockEvents)
		assert.Len(t, matches, 1)
		assert.Equal(t, "Tornado Warning", matches[0].Event)
		assert.Equal(t, "urn:oid:1", matches[0].ID)
	})

	t.Run("wrong zone not matched", func(t *testing.T) {
		loc := mockLocation
		loc.Zone = "FLZ999"
		loc.County = "FLC999"
		matches := FilterAlerts(mockAlertResponse, loc, mockEvents)
		assert.Len(t, matches, 0)
	})

	t.Run("wrong event not matched", func(t *testing.T) {
		matches := FilterAlerts(mockAlertResponse, mockLocation, []string{"Hurricane Warning"})
		assert.Len(t, matches, 0)
	})

	t.Run("multiple events in config", func(t *testing.T) {
		matches := FilterAlerts(mockAlertResponse, mockLocation, []string{"Tornado Warning", "Heat Advisory"})
		assert.Len(t, matches, 2)
	})

	t.Run("empty alerts returns no matches", func(t *testing.T) {
		matches := FilterAlerts(AlertResponse{}, mockLocation, mockEvents)
		assert.Len(t, matches, 0)
	})

	t.Run("correct zone wrong event not matched", func(t *testing.T) {
		matches := FilterAlerts(mockAlertResponse, mockLocation, []string{"Hurricane Warning"})
		assert.Len(t, matches, 0)
	})

	t.Run("matches on county code", func(t *testing.T) {
		loc := mockLocation
		loc.Zone = "FLZ999"
		loc.County = "FLZ112"
		matches := FilterAlerts(mockAlertResponse, loc, mockEvents)
		assert.Len(t, matches, 1)
	})
}

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

		alerts, err := ConnectNOAA(server.Client(), server.URL+"/alerts/", mockLocation, false)
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

		_, err := ConnectNOAA(server.Client(), server.URL+"/alerts/", mockLocation, false)
		assert.Error(t, err)
	})

	t.Run("malformed json", func(t *testing.T) {
		t.Setenv("WEATHERWATCH_USER_AGENT", "weatherwatch (test@example.com)")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid`))
		}))
		defer server.Close()

		_, err := ConnectNOAA(server.Client(), server.URL+"/alerts/", mockLocation, false)
		assert.Error(t, err)
	})

	t.Run("missing user agent", func(t *testing.T) {
		t.Setenv("WEATHERWATCH_USER_AGENT", "")

		_, err := ConnectNOAA(http.DefaultClient, "https://example.com/alerts/", mockLocation, false)
		assert.Error(t, err)
	})
}

func TestEffectiveExpiry(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	furtherFuture := now.Add(48 * time.Hour)

	t.Run("returns Ends when it is after Expires", func(t *testing.T) {
		p := AlertProperties{
			Expires: future,
			Ends:    furtherFuture,
		}
		assert.Equal(t, furtherFuture, EffectiveExpiry(p))
	})

	t.Run("returns Expires when Ends is zero", func(t *testing.T) {
		p := AlertProperties{
			Expires: future,
			Ends:    time.Time{},
		}
		assert.Equal(t, future, EffectiveExpiry(p))
	})

	t.Run("returns Expires when Ends is before Expires", func(t *testing.T) {
		p := AlertProperties{
			Expires: furtherFuture,
			Ends:    future,
		}
		assert.Equal(t, furtherFuture, EffectiveExpiry(p))
	})

	t.Run("returns Expires when Ends equals Expires", func(t *testing.T) {
		p := AlertProperties{
			Expires: future,
			Ends:    future,
		}
		assert.Equal(t, future, EffectiveExpiry(p))
	})
}

func TestShouldNotify(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	furtherFuture := now.Add(48 * time.Hour)

	t.Run("new alert not in seen should notify", func(t *testing.T) {
		seen := SeenAlerts{}
		assert.True(t, ShouldNotify(seen, "Home.KTAE.RP.S.0044", future))
	})

	t.Run("already seen with same expiry should not notify", func(t *testing.T) {
		seen := SeenAlerts{
			"Home.KTAE.RP.S.0044": future,
		}
		assert.False(t, ShouldNotify(seen, "Home.KTAE.RP.S.0044", future))
	})

	t.Run("already seen with earlier expiry should not notify", func(t *testing.T) {
		seen := SeenAlerts{
			"Home.KTAE.RP.S.0044": furtherFuture,
		}
		assert.False(t, ShouldNotify(seen, "Home.KTAE.RP.S.0044", future))
	})

	t.Run("already seen but expiry extended should notify", func(t *testing.T) {
		seen := SeenAlerts{
			"Home.KTAE.RP.S.0044": future,
		}
		assert.True(t, ShouldNotify(seen, "Home.KTAE.RP.S.0044", furtherFuture))
	})

	t.Run("different key not in seen should notify", func(t *testing.T) {
		seen := SeenAlerts{
			"Home.KTAE.RP.S.0044": future,
		}
		assert.True(t, ShouldNotify(seen, "Home.KTAE.TO.W.0001", future))
	})
}

func TestPointInBoundingBox(t *testing.T) {
	// Polygon roughly covering Panama City Beach area
	coordinates := [][]float64{
		{-86.00, 30.27},
		{-85.99, 30.39},
		{-85.96, 30.44},
		{-85.90, 30.44},
		{-85.86, 30.50},
		{-85.52, 30.34},
		{-85.51, 29.96},
		{-86.00, 30.27},
	}

	t.Run("point inside bounding box", func(t *testing.T) {
		// Panama City Beach roughly 30.19, -85.81 — inside the polygon above
		result := pointInBoundingBox(30.19, -85.81, coordinates)
		assert.True(t, result)
	})

	t.Run("point outside bounding box to the north", func(t *testing.T) {
		result := pointInBoundingBox(31.00, -85.81, coordinates)
		assert.False(t, result)
	})

	t.Run("point outside bounding box to the south", func(t *testing.T) {
		result := pointInBoundingBox(29.00, -85.81, coordinates)
		assert.False(t, result)
	})

	t.Run("point outside bounding box to the east", func(t *testing.T) {
		result := pointInBoundingBox(30.19, -84.00, coordinates)
		assert.False(t, result)
	})

	t.Run("point outside bounding box to the west", func(t *testing.T) {
		result := pointInBoundingBox(30.19, -87.00, coordinates)
		assert.False(t, result)
	})

	t.Run("point on bounding box edge", func(t *testing.T) {
		// Exactly on the minimum lat boundary
		result := pointInBoundingBox(29.96, -85.81, coordinates)
		assert.True(t, result)
	})

	t.Run("no geometry coordinates returns false", func(t *testing.T) {
		result := pointInBoundingBox(30.19, -85.81, [][]float64{})
		assert.False(t, result)
	})
}
