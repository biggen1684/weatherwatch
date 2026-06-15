package weather

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
