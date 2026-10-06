package compono

import (
	"bytes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hookCall struct {
	builtin string
	chain   []Frame
}

type recordingAttributeHook struct {
	calls  []hookCall
	result map[string]string
}

func newRecordingAttributeHook(result map[string]string) *recordingAttributeHook {
	return &recordingAttributeHook{result: result}
}

func (h *recordingAttributeHook) hook() AttributeHookFunc {
	return func(builtin string, chain []Frame) map[string]string {
		h.calls = append(h.calls, hookCall{
			builtin: builtin,
			chain:   deepCopyFrames(chain),
		})
		return h.result
	}
}

func deepCopyFrames(chain []Frame) []Frame {
	if chain == nil {
		return nil
	}
	cloned := make([]Frame, len(chain))
	for i, frame := range chain {
		var scopePath []string
		if frame.ScopePath != nil {
			scopePath = make([]string, len(frame.ScopePath))
			copy(scopePath, frame.ScopePath)
		}
		cloned[i] = Frame{
			Name:      frame.Name,
			Kind:      frame.Kind,
			ScopePath: scopePath,
		}
	}
	return cloned
}

func hookBuiltinNames(calls []hookCall) []string {
	names := make([]string, len(calls))
	for i, call := range calls {
		names[i] = call.builtin
	}
	return names
}

func (s *componoTestSuite) convertWithOpts(source string, opts ...ConvertOption) (string, error) {
	var buf bytes.Buffer
	err := New().Convert([]byte(source), &buf, opts...)
	return buf.String(), err
}

func (s *componoTestSuite) requireErrorCode(err error, code ErrorCode) {
	s.T().Helper()
	require.Error(s.T(), err)
	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), code, compErr.Code)
}

const attrHookImageMediaWithVariants = `{
  url: "https://cdn.example.com/photo.jpg",
  width: 1600,
  height: 900,
  mime-type: "image/jpeg",
  variants: [
    {
      url: "https://cdn.example.com/photo-640.webp",
      width: 640,
      height: 360,
      mime-type: "image/webp"
    }
  ]
}`

const attrHookImageMediaNoVariants = `{
  url: "https://cdn.example.com/photo.jpg",
  width: 1600,
  height: 900,
  mime-type: "image/jpeg",
  variants: []
}`

func (s *componoTestSuite) TestAttributeHookBuiltinLink() {
	hook := newRecordingAttributeHook(map[string]string{"class": "x"})
	out, err := s.convertWithOpts(`{{ LINK text = "t" url = "/a" }}`, WithAttributeHook(hook.hook()))
	require.Nil(s.T(), err)

	assert.Equal(s.T(), `<a href="/a" class="x">t</a>`, out)
	require.Len(s.T(), hook.calls, 1)
	assert.Equal(s.T(), "LINK", hook.calls[0].builtin)
	assert.Len(s.T(), hook.calls[0].chain, 0)
}

func (s *componoTestSuite) TestAttributeHookChainGlobalScope() {
	hook := newRecordingAttributeHook(nil)
	_, err := s.convertWithOpts(
		`{{ UMONO_LAYOUT }}`,
		WithGlobalComponent("UMONO_LAYOUT", []byte(`{{ MAIN_MENU }}`), WithIsolatedScope(),
			WithGlobalComponent("MAIN_MENU", []byte(`{{ LINK text = "t" url = "/a" }}`)),
		),
		WithAttributeHook(hook.hook()),
	)
	require.Nil(s.T(), err)

	require.Len(s.T(), hook.calls, 1)
	assert.Equal(s.T(), []Frame{
		{Name: "UMONO_LAYOUT", Kind: FrameGlobal, ScopePath: []string{}},
		{Name: "MAIN_MENU", Kind: FrameGlobal, ScopePath: []string{"UMONO_LAYOUT"}},
	}, hook.calls[0].chain)
}

func (s *componoTestSuite) TestAttributeHookChainPageLocal() {
	hook := newRecordingAttributeHook(nil)
	_, err := s.convertWithOpts(
		"{{ CARD }}\n\n~ CARD\n{{ LINK text = \"t\" url = \"/a\" }}",
		WithAttributeHook(hook.hook()),
	)
	require.Nil(s.T(), err)

	require.Len(s.T(), hook.calls, 1)
	assert.Equal(s.T(), []Frame{
		{Name: "CARD", Kind: FrameLocal, ScopePath: nil},
	}, hook.calls[0].chain)
}

func (s *componoTestSuite) TestAttributeHookChainLocalInRootGlobal() {
	hook := newRecordingAttributeHook(nil)
	_, err := s.convertWithOpts(
		`{{ G }}`,
		WithGlobalComponent("G", []byte("{{ L }}\n\n~ L\n{{ LINK text = \"t\" url = \"/a\" }}")),
		WithAttributeHook(hook.hook()),
	)
	require.Nil(s.T(), err)

	require.Len(s.T(), hook.calls, 1)
	assert.Equal(s.T(), []Frame{
		{Name: "G", Kind: FrameGlobal, ScopePath: []string{}},
		{Name: "L", Kind: FrameLocal, ScopePath: nil},
	}, hook.calls[0].chain)
}

func (s *componoTestSuite) TestAttributeHookChainDynamicLocalArgs() {
	hook := newRecordingAttributeHook(nil)
	_, err := s.convertWithOpts(
		"{{ WRAP content = BODY }}\n\n~ WRAP content = NO_MATTER\n{{ content }}\n\n~ BODY\n{{ LINK text = \"t\" url = \"/a\" }}",
		WithAttributeHook(hook.hook()),
	)
	require.Nil(s.T(), err)

	require.Len(s.T(), hook.calls, 1)
	assert.Equal(s.T(), []Frame{
		{Name: "WRAP", Kind: FrameLocal, ScopePath: nil},
		{Name: "BODY", Kind: FrameLocal, ScopePath: nil},
	}, hook.calls[0].chain)
}

func (s *componoTestSuite) TestAttributeHookNotCalledForLocalOverride() {
	hook := newRecordingAttributeHook(nil)
	out, err := s.convertWithOpts(
		"{{ LINK }}\n\n~ LINK\nmine",
		WithAttributeHook(hook.hook()),
	)
	require.Nil(s.T(), err)

	assert.Empty(s.T(), hook.calls)
	assert.NotContains(s.T(), out, "<a ")
}

func (s *componoTestSuite) TestAttributeHookNotCalledForMarkdownLink() {
	hook := newRecordingAttributeHook(nil)
	_, err := s.convertWithOpts(`[t](/u)`, WithAttributeHook(hook.hook()))
	require.Nil(s.T(), err)

	assert.Empty(s.T(), hook.calls)
}

func (s *componoTestSuite) TestAttributeHookNotCalledOnValidationError() {
	const source = `{{ LINK text = "t" url = "/a" foo = "x" }}`

	noHook, err := s.convertWithOpts(source)
	require.Nil(s.T(), err)

	hook := newRecordingAttributeHook(nil)
	withHook, err := s.convertWithOpts(source, WithAttributeHook(hook.hook()))
	require.Nil(s.T(), err)

	assert.Equal(s.T(), noHook, withHook)
	assert.Empty(s.T(), hook.calls)
}

func (s *componoTestSuite) TestAttributeHookOrderAndEscaping() {
	hook := newRecordingAttributeHook(map[string]string{
		"data-b": `x"<`,
		"class":  "c",
	})
	out, err := s.convertWithOpts(`{{ LINK text = "t" url = "/a" }}`, WithAttributeHook(hook.hook()))
	require.Nil(s.T(), err)

	assert.Equal(s.T(), `<a href="/a" class="c" data-b="x&#34;&lt;">t</a>`, out)
}

func (s *componoTestSuite) TestAttributeHookNilAndEmptyMatchNoHook() {
	const source = `{{ LINK text = "t" url = "/a" }}`

	noHook, err := s.convertWithOpts(source)
	require.Nil(s.T(), err)

	nilMap, err := s.convertWithOpts(source, WithAttributeHook(newRecordingAttributeHook(nil).hook()))
	require.Nil(s.T(), err)
	assert.Equal(s.T(), noHook, nilMap)

	emptyMap, err := s.convertWithOpts(source, WithAttributeHook(newRecordingAttributeHook(map[string]string{}).hook()))
	require.Nil(s.T(), err)
	assert.Equal(s.T(), noHook, emptyMap)

	noFn, err := s.convertWithOpts(source, WithAttributeHook(nil))
	require.Nil(s.T(), err)
	assert.Equal(s.T(), noHook, noFn)
}

func (s *componoTestSuite) TestAttributeHookRenderOrderOncePerCall() {
	hook := newRecordingAttributeHook(nil)
	source := "{{ LINK text = \"1\" url = \"/1\" }}\n\n" +
		"{{ IMAGE media = " + attrHookImageMediaWithVariants + " alt = \"a\" }}\n\n" +
		"{{ LINK text = \"2\" url = \"/2\" }}"

	_, err := s.convertWithOpts(source, WithAttributeHook(hook.hook()))
	require.Nil(s.T(), err)

	require.Len(s.T(), hook.calls, 3)
	assert.Equal(s.T(), []string{"LINK", "IMAGE", "LINK"}, hookBuiltinNames(hook.calls))
}

func (s *componoTestSuite) TestAttributeHookImageWithVariants() {
	hook := newRecordingAttributeHook(map[string]string{
		"class": "c",
		"sizes": "100vw",
	})
	source := `{{ IMAGE media = ` + attrHookImageMediaWithVariants + ` alt = "Photo" }}`
	out, err := s.convertWithOpts(source, WithAttributeHook(hook.hook()))
	require.Nil(s.T(), err)

	assert.Equal(s.T(),
		`<picture class="c"><source type="image/webp" srcset="https://cdn.example.com/photo-640.webp 640w" sizes="100vw"><img src="https://cdn.example.com/photo.jpg" alt="Photo" width="1600" height="900"></picture>`,
		out,
	)
}

func (s *componoTestSuite) TestAttributeHookImageWithoutVariants() {
	source := `{{ IMAGE media = ` + attrHookImageMediaNoVariants + ` alt = "Photo" }}`

	noHook, err := s.convertWithOpts(source)
	require.Nil(s.T(), err)

	hook := newRecordingAttributeHook(map[string]string{"sizes": "100vw"})
	withHook, err := s.convertWithOpts(source, WithAttributeHook(hook.hook()))
	require.Nil(s.T(), err)

	assert.Equal(s.T(), noHook, withHook)
}

func (s *componoTestSuite) TestAttributeHookAlreadySet() {
	const source = `{{ LINK text = "t" url = "/a" }}`

	_, err := s.convertWithOpts(source, WithAttributeHook(nil), WithAttributeHook(nil))
	s.requireErrorCode(err, ErrAttributeHookAlreadySet)

	hook := newRecordingAttributeHook(nil)
	_, err = s.convertWithOpts(source, WithAttributeHook(hook.hook()), WithAttributeHook(nil))
	s.requireErrorCode(err, ErrAttributeHookAlreadySet)
}

func (s *componoTestSuite) TestAttributeHookInvalidName() {
	const source = `{{ LINK text = "t" url = "/a" }}`

	for _, name := range []string{"Class", "1a", "a b", ""} {
		hook := newRecordingAttributeHook(map[string]string{name: "v"})
		_, err := s.convertWithOpts(source, WithAttributeHook(hook.hook()))
		s.requireErrorCode(err, ErrInvalidAttributeName)
	}
}

func (s *componoTestSuite) TestAttributeHookConflict() {
	const source = `{{ LINK text = "t" url = "/a" }}`

	for _, name := range []string{"href", "target", "rel"} {
		hook := newRecordingAttributeHook(map[string]string{name: "v"})
		_, err := s.convertWithOpts(source, WithAttributeHook(hook.hook()))
		s.requireErrorCode(err, ErrAttributeConflict)
	}
}
