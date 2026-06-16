package weather

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPushoverAPIKey(t *testing.T) {
	t.Run("key is set", func(t *testing.T) {
		t.Setenv("PUSHOVER_API_KEY", "abc123")

		key, err := getPushoverAPIKey()
		assert.NoError(t, err)
		assert.Equal(t, "abc123", key)
	})

	t.Run("key is missing", func(t *testing.T) {
		t.Setenv("PUSHOVER_API_KEY", "")

		_, err := getPushoverAPIKey()
		assert.Error(t, err)
	})
}

func TestGetPushoverUserKey(t *testing.T) {
	t.Run("user key is set", func(t *testing.T) {
		t.Setenv("PUSHOVER_USER_KEY", "xyz789")

		key, err := getPushoverUserKey()
		assert.NoError(t, err)
		assert.Equal(t, "xyz789", key)
	})

	t.Run("user key is missing", func(t *testing.T) {
		t.Setenv("PUSHOVER_USER_KEY", "")

		_, err := getPushoverUserKey()
		assert.Error(t, err)
	})
}

func TestGetUserAgent(t *testing.T) {
	t.Run("user agent is set", func(t *testing.T) {
		t.Setenv("WEATHERWATCH_USER_AGENT", "weatherwatch (test@example.com)")

		ua, err := getUserAgent()
		assert.NoError(t, err)
		assert.Equal(t, "weatherwatch (test@example.com)", ua)
	})

	t.Run("user agent is missing", func(t *testing.T) {
		t.Setenv("WEATHERWATCH_USER_AGENT", "")

		_, err := getUserAgent()
		assert.Error(t, err)
	})
}
