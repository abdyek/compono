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

func TestComponentReference(t *testing.T) {
	t.Run("selects component name", func(t *testing.T) {
		selected := NewComponentReference().Select([]byte(`MAIN_MENU`))
		assert.Equal(t, [][2]int{{0, 9}}, selected)
	})

	t.Run("selects component with bound arguments", func(t *testing.T) {
		selected := NewComponentReference().Select([]byte(`MAIN_MENU(menu = menu label = "Menu")`))
		assert.Equal(t, [][2]int{{0, 37}}, selected)
	})

	t.Run("selects nested bound arguments over multiple lines", func(t *testing.T) {
		selected := NewComponentReference().Select([]byte("X(\n  c = Y(a = 1)\n  items = [Z(b = true)]\n)"))
		assert.Equal(t, [][2]int{{0, 43}}, selected)
	})

	t.Run("rejects empty parenthesis", func(t *testing.T) {
		selected := NewComponentReference().Select([]byte(`X()`))
		assert.Empty(t, selected)
	})

	t.Run("rejects space between name and parenthesis", func(t *testing.T) {
		selected := NewComponentReference().Select([]byte(`X (a = 1)`))
		assert.Empty(t, selected)
	})

	t.Run("rejects invalid bound arguments", func(t *testing.T) {
		selected := NewComponentReference().Select([]byte(`X(a)`))
		assert.Empty(t, selected)
	})
}

func TestBoundArgs(t *testing.T) {
	t.Run("selects parenthesized arguments", func(t *testing.T) {
		selected := NewBoundArgs().Select([]byte(`MAIN_MENU(menu = menu)`))
		assert.Equal(t, [][2]int{{9, 22}}, selected)
	})

	t.Run("selects nothing without arguments", func(t *testing.T) {
		selected := NewBoundArgs().Select([]byte(`MAIN_MENU`))
		assert.Empty(t, selected)
	})
}

func TestComponentAssignmentsWithBoundArgs(t *testing.T) {
	selected := NewComponentAssignments(true).Select([]byte(`content = MAIN_MENU(menu = menu) label = "Menu"`))
	assert.Equal(t, [][2]int{{0, 32}, {33, 47}}, selected)
}
