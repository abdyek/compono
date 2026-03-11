package selector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayLiteral(t *testing.T) {
	t.Run("rejects param refs when disabled", func(t *testing.T) {
		selected := NewArrayLiteral(false).Select([]byte(`[param]`))
		assert.Empty(t, selected)
	})

	t.Run("accepts param refs when enabled", func(t *testing.T) {
		selected := NewArrayLiteral(true).Select([]byte(`[param]`))
		assert.Equal(t, [][2]int{{0, 7}}, selected)
	})
}

func TestRecordLiteral(t *testing.T) {
	t.Run("rejects param refs when disabled", func(t *testing.T) {
		selected := NewRecordLiteral(false).Select([]byte(`{key: param}`))
		assert.Empty(t, selected)
	})

	t.Run("accepts param refs and nested access when enabled", func(t *testing.T) {
		selected := NewRecordLiteral(true).Select([]byte(`{key: nested.value[0]}`))
		assert.Equal(t, [][2]int{{0, 22}}, selected)
	})
}
