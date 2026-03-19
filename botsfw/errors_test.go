package botsfw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrAuthFailed_Error(t *testing.T) {
	t.Run("returns_string_value", func(t *testing.T) {
		err := ErrAuthFailed("authentication failed")
		assert.Equal(t, "authentication failed", err.Error())
	})

	t.Run("empty_string", func(t *testing.T) {
		err := ErrAuthFailed("")
		assert.Equal(t, "", err.Error())
	})

	t.Run("implements_error_interface", func(t *testing.T) {
		var err error = ErrAuthFailed("test")
		assert.NotNil(t, err)
		assert.Equal(t, "test", err.Error())
	})
}

func TestErrEntityNotFound(t *testing.T) {
	assert.NotNil(t, ErrEntityNotFound)
	assert.Equal(t, "bots-framework: no such entity", ErrEntityNotFound.Error())
}
