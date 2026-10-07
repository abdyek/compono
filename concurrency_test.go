package compono

import (
	"bytes"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func caseOpts(i int) ([]byte, []ConvertOption) {
	v := strconv.Itoa(i)
	src := "# {{ context(title) }}\n\n{{ PAGE }}\n\n{{ LINK text = \"page\" url = \"/p" + v + "\" }}"
	if i%2 == 1 {
		src += "\n\n{{ MISSING }}"
	}

	opts := []ConvertOption{
		WithContext(map[string]any{"title": "Title " + v}),
		WithGlobalComponent("CARD", []byte("root card "+v)),
		WithGlobalComponent("PAGE", []byte("{{ CARD }}"), WithIsolatedScope(), WithGlobalComponent("CARD", []byte("{{ LINK text = \"card"+v+"\" url = \"/c"+v+"\" }}"))),
		WithAttributeHook(func(builtin string, stack []Frame) map[string]string {
			return map[string]string{"data-case": v, "data-depth": strconv.Itoa(len(stack))}
		}),
	}

	return []byte(src), opts
}

func TestConcurrentConvert(t *testing.T) {
	const n = 16

	expected := make([][]byte, n)
	for i := 0; i < n; i++ {
		v := strconv.Itoa(i)
		src, opts := caseOpts(i)

		var buf bytes.Buffer
		require.NoError(t, New().Convert(src, &buf, opts...))
		expected[i] = buf.Bytes()

		assert.Contains(t, string(expected[i]), "Title "+v)
		assert.Contains(t, string(expected[i]), "/p"+v)
		assert.Contains(t, string(expected[i]), "card"+v)
	}

	shared := New()

	const goroutines = n * 8
	outputs := make([][]byte, goroutines)
	errs := make([]error, goroutines)

	var wg sync.WaitGroup
	for k := 0; k < goroutines; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()

			src, opts := caseOpts(k % n)

			var buf bytes.Buffer
			errs[k] = shared.Convert(src, &buf, opts...)
			outputs[k] = buf.Bytes()
		}(k)
	}
	wg.Wait()

	for k := 0; k < goroutines; k++ {
		require.NoError(t, errs[k])
		assert.Equal(t, expected[k%n], outputs[k])
	}
}
