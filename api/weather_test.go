package weather

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const mockResponse = `{
"@context": [
        "https://geojson.org/geojson-ld/geojson-context.jsonld",
        {
            "@version": "1.1",
            "wx": "https://api.weather.gov/ontology#",
            "@vocab": "https://api.weather.gov/ontology#"
        }
    ],
    "type": "FeatureCollection",
    "features": [
        {
            "id": "https://api.weather.gov/alerts/urn:oid:2.49.0.1.840.0.a0969f2ced33c4b5fa96093599749247d57ecd23.001.1",
            "type": "Feature",
            "geometry": null,
            "properties": {
                "@id": "https://api.weather.gov/alerts/urn:oid:2.49.0.1.840.0.a0969f2ced33c4b5fa96093599749247d57ecd23.001.1",
                "@type": "wx:Alert",
                "id": "urn:oid:2.49.0.1.840.0.a0969f2ced33c4b5fa96093599749247d57ecd23.001.1",
                "areaDesc": "South Walton; Coastal Bay; Coastal Gulf",
                "geocode": {
                    "SAME": [
                        "012131",
                        "012005",
                        "012045"
                    ],
                    "UGC": [
                        "FLZ108",
                        "FLZ112",
                        "FLZ114"
                    ]
                },
                "affectedZones": [
                    "https://api.weather.gov/zones/forecast/FLZ108",
                    "https://api.weather.gov/zones/forecast/FLZ112",
                    "https://api.weather.gov/zones/forecast/FLZ114"
                ],
                "references": [],
                "sent": "2026-06-13T03:16:00-04:00",
                "effective": "2026-06-13T03:16:00-04:00",
                "onset": "2026-06-13T03:16:00-04:00",
                "expires": "2026-06-14T04:30:00-04:00",
                "ends": "2026-06-15T05:00:00-04:00",
                "status": "Actual",
                "messageType": "Alert",
                "category": "Met",
                "severity": "Moderate",
                "certainty": "Likely",
                "urgency": "Expected",
                "event": "Rip Current Statement",
                "sender": "w-nws.webmaster@noaa.gov",
                "senderName": "NWS Tallahassee FL",
                "headline": "Rip Current Statement issued June 13 at 3:16AM EDT until June 15 at 5:00AM EDT by NWS Tallahassee FL",
                "description": "* WHAT...Dangerous rip currents.\n\n* WHERE...Walton, Bay, and West-facing Gulf County Beaches.\n\n* WHEN...Through late Sunday night.\n\n* IMPACTS...Rip currents can sweep even the best swimmers away\nfrom shore into deeper water.",
                "instruction": "Swim near a lifeguard. If caught in a rip current, relax and\nfloat. Don't swim against the current. If able, swim in a\ndirection following the shoreline. If unable to escape, face the\nshore and call or wave for help.",
                "response": "Avoid",
                "note": null,
                "parameters": {
                    "AWIPSidentifier": [
                        "CFWTAE"
                    ],
                    "WMOidentifier": [
                        "WHUS42 KTAE 130716"
                    ],
                    "NWSheadline": [
                        "HIGH RIP CURRENT RISK IN EFFECT THROUGH LATE SUNDAY NIGHT"
                    ],
                    "BLOCKCHANNEL": [
                        "EAS",
                        "NWEM",
                        "CMAS"
                    ],
                    "VTEC": [
                        "/O.NEW.KTAE.RP.S.0044.260613T0716Z-260615T0900Z/"
                    ],
                    "eventEndingTime": [
                        "2026-06-15T05:00:00-04:00"
                    ]
                },
                "scope": "Public",
                "code": "IPAWSv1.0",
                "language": "en-US",
                "web": "http://www.weather.gov",
                "eventCode": {
                    "SAME": [
                        "NWS"
                    ],
                    "NationalWeatherService": [
                        "RPS"
                    ]
                }
            }
        },
        {
            "id": "https://api.weather.gov/alerts/urn:oid:2.49.0.1.840.0.ae22f1e4ad2243751f2f1cad4a34a57df9c762a1.001.1",
            "type": "Feature",
            "geometry": null,
            "properties": {
                "@id": "https://api.weather.gov/alerts/urn:oid:2.49.0.1.840.0.ae22f1e4ad2243751f2f1cad4a34a57df9c762a1.001.1",
                "@type": "wx:Alert",
                "id": "urn:oid:2.49.0.1.840.0.ae22f1e4ad2243751f2f1cad4a34a57df9c762a1.001.1",
                "areaDesc": "Holmes; Washington; Jackson; Inland Bay; Calhoun; Inland Gulf; Inland Franklin; Gadsden; Leon; Inland Jefferson; Madison; Inland Wakulla; Inland Taylor; Lafayette; Inland Dixie; Coastal Bay; Coastal Gulf; Coastal Franklin; Coastal Jefferson; Coastal Wakulla; Coastal Taylor; Coastal Dixie; Northern Liberty; Southern Liberty; Seminole; Decatur; Grady; Thomas; Brooks; Lowndes; Lanier",
                "geocode": {
                    "SAME": [
                        "012059",
                        "012133",
                        "012063",
                        "012005",
                        "012013",
                        "012045",
                        "012037",
                        "012039",
                        "012073",
                        "012065",
                        "012079",
                        "012129",
                        "012123",
                        "012067",
                        "012029",
                        "012077",
                        "013253",
                        "013087",
                        "013131",
                        "013275",
                        "013027",
                        "013185",
                        "013173"
                    ],
                    "UGC": [
                        "FLZ009",
                        "FLZ010",
                        "FLZ011",
                        "FLZ012",
                        "FLZ013",
                        "FLZ014",
                        "FLZ015",
                        "FLZ016",
                        "FLZ017",
                        "FLZ018",
                        "FLZ019",
                        "FLZ027",
                        "FLZ028",
                        "FLZ029",
                        "FLZ034",
                        "FLZ112",
                        "FLZ114",
                        "FLZ115",
                        "FLZ118",
                        "FLZ127",
                        "FLZ128",
                        "FLZ134",
                        "FLZ326",
                        "FLZ426",
                        "GAZ155",
                        "GAZ156",
                        "GAZ157",
                        "GAZ158",
                        "GAZ159",
                        "GAZ160",
                        "GAZ161"
                    ]
                },
                "affectedZones": [
                    "https://api.weather.gov/zones/forecast/FLZ009",
                    "https://api.weather.gov/zones/forecast/FLZ010",
                    "https://api.weather.gov/zones/forecast/FLZ011",
                    "https://api.weather.gov/zones/forecast/FLZ012",
                    "https://api.weather.gov/zones/forecast/FLZ013",
                    "https://api.weather.gov/zones/forecast/FLZ014",
                    "https://api.weather.gov/zones/forecast/FLZ015",
                    "https://api.weather.gov/zones/forecast/FLZ016",
                    "https://api.weather.gov/zones/forecast/FLZ017",
                    "https://api.weather.gov/zones/forecast/FLZ018",
                    "https://api.weather.gov/zones/forecast/FLZ019",
                    "https://api.weather.gov/zones/forecast/FLZ027",
                    "https://api.weather.gov/zones/forecast/FLZ028",
                    "https://api.weather.gov/zones/forecast/FLZ029",
                    "https://api.weather.gov/zones/forecast/FLZ034",
                    "https://api.weather.gov/zones/forecast/FLZ112",
                    "https://api.weather.gov/zones/forecast/FLZ114",
                    "https://api.weather.gov/zones/forecast/FLZ115",
                    "https://api.weather.gov/zones/forecast/FLZ118",
                    "https://api.weather.gov/zones/forecast/FLZ127",
                    "https://api.weather.gov/zones/forecast/FLZ128",
                    "https://api.weather.gov/zones/forecast/FLZ134",
                    "https://api.weather.gov/zones/forecast/FLZ326",
                    "https://api.weather.gov/zones/forecast/FLZ426",
                    "https://api.weather.gov/zones/forecast/GAZ155",
                    "https://api.weather.gov/zones/forecast/GAZ156",
                    "https://api.weather.gov/zones/forecast/GAZ157",
                    "https://api.weather.gov/zones/forecast/GAZ158",
                    "https://api.weather.gov/zones/forecast/GAZ159",
                    "https://api.weather.gov/zones/forecast/GAZ160",
                    "https://api.weather.gov/zones/forecast/GAZ161"
                ],
                "references": [
                    {
                        "@id": "https://api.weather.gov/alerts/urn:oid:2.49.0.1.840.0.4d45958b66c0681fe0ff8a5ca64a858158522c6e.001.1",
                        "identifier": "urn:oid:2.49.0.1.840.0.4d45958b66c0681fe0ff8a5ca64a858158522c6e.001.1",
                        "sender": "w-nws.webmaster@noaa.gov",
                        "sent": "2026-06-12T13:00:00-04:00"
                    }
                ],
                "sent": "2026-06-13T02:26:00-04:00",
                "effective": "2026-06-13T02:26:00-04:00",
                "onset": "2026-06-13T12:00:00-04:00",
                "expires": "2026-06-13T18:00:00-04:00",
                "ends": "2026-06-13T18:00:00-04:00",
                "status": "Actual",
                "messageType": "Update",
                "category": "Met",
                "severity": "Moderate",
                "certainty": "Likely",
                "urgency": "Expected",
                "event": "Heat Advisory",
                "sender": "w-nws.webmaster@noaa.gov",
                "senderName": "NWS Tallahassee FL",
                "headline": "Heat Advisory issued June 13 at 2:26AM EDT until June 13 at 6:00PM EDT by NWS Tallahassee FL",
                "description": "* WHAT...Heat index values of 106 to 110 degrees expected.\n\n* WHERE...Portions of the Eastern Florida Panhandle, Florida Big\nBend, and far Southwestern Georgia.\n\n* WHEN...From noon EDT /11 AM CDT/ today to 6 PM EDT /5 PM CDT/ this\nevening.\n\n* IMPACTS...Hot temperatures and high humidity may increase\npotential for heat-related illnesses.",
                "instruction": "Drink plenty of fluids, stay in an air-conditioned room, stay out of\nthe sunshine, and check up on relatives and neighbors.\n\nTo reduce risk during outdoor work, the Occupational Safety and\nHealth Administration recommends scheduling frequent rest breaks in\nshaded or air conditioned environments. Anyone overcome by heat\nshould be moved to a cool and shaded location. Heat stroke is an\nemergency! Call 9 1 1.",
                "response": "Execute",
                "note": null,
                "parameters": {
                    "AWIPSidentifier": [
                        "NPWTAE"
                    ],
                    "WMOidentifier": [
                        "WWUS72 KTAE 130626"
                    ],
                    "NWSheadline": [
                        "HEAT ADVISORY REMAINS IN EFFECT FROM NOON EDT /11 AM CDT/ TODAY TO 6 PM EDT /5 PM CDT/ THIS EVENING"
                    ],
                    "BLOCKCHANNEL": [
                        "EAS",
                        "NWEM",
                        "CMAS"
                    ],
                    "VTEC": [
                        "/O.CON.KTAE.HT.Y.0001.260613T1600Z-260613T2200Z/"
                    ],
                    "eventEndingTime": [
                        "2026-06-13T18:00:00-04:00"
                    ]
                },
                "scope": "Public",
                "code": "IPAWSv1.0",
                "language": "en-US",
                "web": "http://www.weather.gov",
                "eventCode": {
                    "SAME": [
                        "NWS"
                    ],
                    "NationalWeatherService": [
                        "HTY"
                    ]
                }
            }
        }
    ],
    "title": "Current watches, warnings, and advisories for Florida",
    "updated": "2026-06-13T14:28:18+00:00"
}`

func TestGetPushoverKey(t *testing.T) {
	t.Run("key is set", func(t *testing.T) {
		t.Setenv("PUSHOVER_API_KEY", "abc123")

		key, err := getPushoverKey()
		assert.NoError(t, err)
		assert.Equal(t, "abc123", key)
	})

	t.Run("key is missing", func(t *testing.T) {
		t.Setenv("PUSHOVER_API_KEY", "")

		_, err := getPushoverKey()
		assert.Error(t, err)
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("valid toml file", func(t *testing.T) {
		content := `
office = "NWS Tallahassee FL"
area = "FL"
user_agent = "weatherwatch (test@example.com)"
events = ["Tornado Warning"]
`
		path := filepath.Join(t.TempDir(), "config.toml")
		err := os.WriteFile(path, []byte(content), 0644)
		assert.NoError(t, err)

		cfg, err := loadConfig(path)
		assert.NoError(t, err)
		assert.Equal(t, "NWS Tallahassee FL", cfg.Office)
		assert.Equal(t, "FL", cfg.Area)
		assert.Equal(t, []string{"Tornado Warning"}, cfg.Events)
	})

	t.Run("file does not exist", func(t *testing.T) {
		_, err := loadConfig("nonexistent.toml")
		assert.Error(t, err)
	})

	t.Run("malformed toml", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bad.toml")
		err := os.WriteFile(path, []byte("office = ["), 0644)
		assert.NoError(t, err)

		_, err = loadConfig(path)
		assert.Error(t, err)
	})
}

func TestValidateConfig(t *testing.T) {
	valid := Config{
		Office:    "NWS Tallahassee FL",
		Area:      "FL",
		UserAgent: "weatherwatch (test@example.com)",
		Events:    []string{"Tornado Warning"},
	}

	t.Run("valid config", func(t *testing.T) {
		err := validateConfig(valid)
		assert.NoError(t, err)
	})

	t.Run("missing area", func(t *testing.T) {
		cfg := valid
		cfg.Area = ""
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("missing events", func(t *testing.T) {
		cfg := valid
		cfg.Events = nil
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("missing office", func(t *testing.T) {
		cfg := valid
		cfg.Office = ""
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("missing user agent", func(t *testing.T) {
		cfg := valid
		cfg.UserAgent = ""
		err := validateConfig(cfg)
		assert.Error(t, err)
	})
}

// func TestConnectNOAA_Success(t *testing.T) {
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte(mockResponse))
// 	}))
// 	defer server.Close()

// 	client := server.Client()
// 	alerts, err := ConnectNOAA(client, server.URL+"/", false)

// 	assert.NoError(t, err)
// 	assert.Equal(t, 2, len(alerts.Features))
// 	assert.Equal(t, "Rip Current Statement", alerts.Features[0].Properties.Event)
// 	assert.Equal(t, "NWS Tallahassee FL", alerts.Features[0].Properties.SenderName)
// }

// func TestConnectNOAA_NonOKStatus(t *testing.T) {
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusInternalServerError)
// 	}))
// 	defer server.Close()

// 	_, err := ConnectNOAA(server.Client(), server.URL+"/", false)
// 	assert.Error(t, err)
// }

// func TestConnectNOAA_BadJSON(t *testing.T) {
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte(`{invalid json`))
// 	}))
// 	defer server.Close()

// 	_, err := ConnectNOAA(server.Client(), server.URL+"/", false)
// 	assert.Error(t, err)
//}
