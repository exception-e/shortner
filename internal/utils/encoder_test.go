package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeBase62(t *testing.T) {
	var testUrlHash uint64 = 916777411

	t.Run("should be alphanumeric", func(t *testing.T) {
		result := EncodeBase62(testUrlHash)
		assert.Regexp(t, "^[a-zA-Z0-9]+$", result, "should be alphanumeric")
	})

	t.Run("should be deterministic", func(t *testing.T) {
		result1 := EncodeBase62(testUrlHash)
		result2 := EncodeBase62(testUrlHash)
		assert.Equal(t, result1, result2, "should return same alias for same input")
	})

	t.Run("should return different alias for different input", func(t *testing.T) {
		result1 := EncodeBase62(testUrlHash)
		result2 := EncodeBase62(123456789)
		assert.NotEqual(t, result1, result2, "should return different alias for different input")
	})
	t.Run("should return 0 for 0 input", func(t *testing.T) {
		result := EncodeBase62(0)
		assert.Equal(t, "0", result)
	})
}

func TestEncodeBase62WithGetHash(t *testing.T) {
	testUrl := "https://www.google.com/search?q=%D1%8E%D0%BD%D0%B8%D1%82+%D1%82%D0%B5%D1%81%D1%82%D1%8B+go&hl=ru&sca_esv=58f7955badf9fa4a&sxsrf=APpeQnvyw9m4EYjbZm-y7h7NhKKAH8Szmg%3A1788943924415&ei=NB6hat75GLLuwPAPxrLAkA8&biw=1330&bih=770&ved=2ahUKEwje44y5j-GWAxUyNxAIHUYZEPIQ4dUDegQIBhAM&uact=5&oq=%D1%8E%D0%BD%D0%B8%D1%82+%D1%82%D0%B5%D1%81%D1%82%D1%8B+go&gs_lp=Egxnd3Mtd2l6LXNlcnAiFtGO0L3QuNGCINGC0LXRgdGC0YsgZ28yBhAAGBYYHjIGEAAYFhgeMgYQABgWGB4yBhAAGBYYHjIGEAAYFhgeMgYQABgWGB4yBhAAGBYYHjIGEAAYFhgeMgYQABgWGB4yBhAAGBYYHkiuhQFQ7BdYp3hwBXgBkAEAmAFEoAHDBKoBAjEyuAEDyAEA-AEBmAIRoAK6BagCCsICChAAGEcY1gQYsAPCAg4QABjkAhjWBBiwA9gBAcICFxAuGNwGGLgGGNoGGNgCGMgDGLAD2AEBwgIKEAAYgAQYigUYQ8ICChAuGEMYgAQYigXCAg0QABiABBiKBRhDGLQHwgIKEC4YgAQYigUYQ8ICERAuGIAEGLEDGIMBGMcBGNEDwgIOEAAYgAQYigUYsQMYgwHCAggQABiABBixA8ICCxAAGIAEGLEDGIMBwgINEAAYgAQYigUYQxixA8ICGRAuGIAEGIoFGEMYlwUY3AQY3gQY3wTYAQHCAhAQABgDGI8BGOoCGLQC2AECwgIQEC4YAxiPARjqAhi0AtgBAsICCxAuGIAEGLEDGIMBwgIIEC4YgAQYsQPCAgsQLhiABBjHARivAcICGhAuGIAEGLEDGIMBGJcFGNwEGN4EGOAE2AEBwgIFEAAYgATCAg4QLhiABBixAxjHARjRA8ICDhAuGIAEGMcBGK8BGI4FwgIOEC4YxwEYsQMY0QMYgATCAhAQLhiABBjHARivARiOBRgKwgIHEAAYgAQYCsICCBAAGIAEGLQHwgIIEAAYiQUYogTCAgUQABjvBcICBhAAGB4YDcICCBAAGBYYHhgKmAMI8QXIus67KJkne4gGAZAGDboGBggBEAEYCboGBAgCGAqSBwIxN6AHyHeyBwIxMrgHlwXCBwQyLTE3yAdXgAgB&sclient=gws-wiz-serp"
	testUrl2 := "https://www.google.com"

	t.Run("should be alphanumeric", func(t *testing.T) {
		result := EncodeBase62(GetHash(testUrl))
		assert.Regexp(t, "^[a-zA-Z0-9]+$", result, "should be alphanumeric")
	})

	t.Run("should be deterministic", func(t *testing.T) {
		result1 := EncodeBase62(GetHash(testUrl))
		result2 := EncodeBase62(GetHash(testUrl))
		assert.Equal(t, result1, result2, "should return same alias for same input")
	})

	t.Run("should return different alias for different input", func(t *testing.T) {
		result1 := EncodeBase62(GetHash(testUrl))
		result2 := EncodeBase62(GetHash(testUrl2))
		assert.NotEqual(t, result1, result2, "should return different alias for different input")
	})
	t.Run("should work with empty input", func(t *testing.T) {
		result := EncodeBase62(GetHash(""))
		assert.NotEmpty(t, result)
	})
}
