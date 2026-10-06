package attrhook

import "errors"

// FrameKind identifies the origin of a Frame on the render stack.
type FrameKind string

const (
	// FrameBuiltin marks a built-in component frame.
	FrameBuiltin FrameKind = "builtin"
	// FrameGlobal marks a frame produced by a global component definition.
	FrameGlobal FrameKind = "global"
	// FrameLocal marks a frame produced by a local component definition.
	FrameLocal FrameKind = "local"
)

// Frame is a single component definition on the render stack. Name is the
// component name, Kind its origin and ScopePath, for global frames, the names
// of the enclosing global component definitions ordered outermost first. For
// root globals ScopePath is a non-nil empty slice; for local frames it is nil.
type Frame struct {
	Name      string
	Kind      FrameKind
	ScopePath []string
}

// AttributeHookFunc is invoked once for every built-in component render. It
// receives the built-in name and the current frame stack and returns the extra
// HTML attributes to write on the rendered element.
type AttributeHookFunc func(builtin string, stack []Frame) map[string]string

var (
	// ErrInvalidAttributeName reports an attribute hook returning a name that
	// does not match ^[a-z][a-z0-9-]*$.
	ErrInvalidAttributeName = errors.New("invalid attribute name")
	// ErrAttributeConflict reports an attribute hook returning a name that
	// Compono already writes on the target element.
	ErrAttributeConflict = errors.New("attribute conflicts with a built-in attribute")
)
