package compono

import (
	"bytes"
	"testing"

	"github.com/umono-cms/compono"
)

func TestDummy(t *testing.T) {
	comp := compono.New()

	comp.RegisterGlobalComponent("SAY_HELLO", []byte(`{{ ANOTHER_R content="selam" }}

~ ANOTHER_R
content = ""

$content
`))

	var buf bytes.Buffer
	comp.Convert([]byte("{{ SAY_HELLO }}"), &buf)
}
