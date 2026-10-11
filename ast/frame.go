package ast

import "github.com/umono-cms/compono/rule"

// compValueFrame is a synthetic frame for a component value rendered without a
// call of its own, such as the content of a built-in.
type compValueFrame struct {
	node
	value ResolvedValue
}

func NewCompValueFrame(value ResolvedValue, parent Node) Node {
	frame := &compValueFrame{value: value}
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
