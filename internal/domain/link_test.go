package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLink(t *testing.T) {

	expectedLink := &Link{
		Alias:       "3XqGtZ",
		OriginalURL: "https://google.com",
	}

	t.Run("create new link successfully", func(t *testing.T) {
		result, err := NewLink("https://google.com", "3XqGtZ")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expectedLink.Alias, result.Alias)
		assert.Equal(t, expectedLink.OriginalURL, result.OriginalURL)
	})

	t.Run("create new link with empty original url fails ", func(t *testing.T) {
		result, err := NewLink("", "3XqGtZ")
		assert.Nil(t, result)
		assert.EqualError(t, err, "domain: original URL cannot be empty")
	})

	t.Run("create new link with empty alias fails ", func(t *testing.T) {
		result, err := NewLink("https://google.com", "")
		assert.Empty(t, result)
		assert.EqualError(t, err, "domain: short code cannot be empty")
	})

	t.Run("create new link with invalid original url fails ", func(t *testing.T) {
		result, err := NewLink("https://[google.com]", "3XqGtZ")
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "domain: invalid URL:")
	})
}
