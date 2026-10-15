package ast

import (
	"sort"

	"github.com/umono-cms/compono/rule"
)

// compValueFrame is a synthetic frame for a component value rendered without a
// call of its own, such as the content of a built-in.
type compValueFrame struct {
	node
	value ResolvedValue
	args  map[string]ResolvedValue
}

func NewCompValueFrame(value ResolvedValue, parent Node) Node {
	return NewCompValueFrameWithArgs(value, parent, nil)
}

// NewCompValueFrameWithArgs is NewCompValueFrame with arguments given to the
// component value by the frame itself, such as the item that LIST gives to its
// item component.
func NewCompValueFrameWithArgs(value ResolvedValue, parent Node, args map[string]ResolvedValue) Node {
	frame := &compValueFrame{value: value, args: args}
	frame.SetRule(rule.NewDynamic("comp-value-frame"))
	frame.SetParent(parent)
	frame.SetRange(parent.Range())
	return frame
}

func GetCompValueFromFrame(node Node) (ResolvedValue, bool) {
	frame, ok := node.(*compValueFrame)
	if !ok {
		return ResolvedValue{}, false
	}
	return frame.value, true
}

// GetCompValueFrameArg returns the argument the comp-value-frame gives to its component value.
func GetCompValueFrameArg(node Node, name string) (ResolvedValue, bool) {
	frame, ok := node.(*compValueFrame)
	if !ok || frame.args == nil {
		return ResolvedValue{}, false
	}
	value, ok := frame.args[name]
	return value, ok
}

// GetCompValueFrameArgNames returns the sorted names of the arguments given
// to the component value by the comp-value-frame node itself, nil if node is
// not such a frame or has no arguments.
func GetCompValueFrameArgNames(node Node) []string {
	frame, ok := node.(*compValueFrame)
	if !ok || frame.args == nil {
		return nil
	}
	names := make([]string, 0, len(frame.args))
	for name := range frame.args {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
