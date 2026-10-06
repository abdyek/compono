package compono

import (
	"fmt"

	"github.com/umono-cms/compono/internal/attrhook"
)

// FrameKind identifies the origin of a Frame on the render stack.
type FrameKind = attrhook.FrameKind

// Frame is a single component definition on the render stack.
type Frame = attrhook.Frame

// AttributeHookFunc is invoked once for every built-in component render.
type AttributeHookFunc = attrhook.AttributeHookFunc

const (
	// FrameBuiltin marks a built-in component frame.
	FrameBuiltin = attrhook.FrameBuiltin
	// FrameGlobal marks a frame produced by a global component definition.
	FrameGlobal = attrhook.FrameGlobal
	// FrameLocal marks a frame produced by a local component definition.
	FrameLocal = attrhook.FrameLocal
)

type attributeHookOption struct {
	fn AttributeHookFunc
}

// WithAttributeHook returns a ConvertOption that registers a function invoked
// once for every built-in component render. The function receives the built-in
// name and the current frame stack and returns extra HTML attributes to write
// on the rendered element. It may be used at most once per conversion and is
// not permitted inside WithGlobalComponent.
func WithAttributeHook(fn AttributeHookFunc) ConvertOption {
	return &attributeHookOption{fn: fn}
}

func (o *attributeHookOption) applyConvert(_ *compono, cfg *convertConfig) error {
	if cfg.attributeHookSet {
		return NewComponoError(ErrAttributeHookAlreadySet, fmt.Sprintf("attribute hook is already set: WithAttributeHook can be used at most once per conversion"))
	}
	cfg.attributeHookSet = true
	cfg.attributeHook = o.fn
	return nil
}
