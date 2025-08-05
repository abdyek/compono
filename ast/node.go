package ast

import "github.com/umono-cms/compono/component"

type Node interface {
	Component() component.Component
	SetComponent(component.Component)
	Parent() Node
	SetParent(Node)
	Children() []Node
	SetChildren([]Node)
	HasChildren() bool
	Raw() []byte
	SetRaw([]byte)
}

func DefaultNode() Node {
	return &node{}
}

type node struct {
	component component.Component
	parent    Node
	children  []Node
	raw       []byte
}

func (n *node) Component() component.Component {
	return n.component
}

func (n *node) SetComponent(comp component.Component) {
	n.component = comp
}

func (n *node) Parent() Node {
	return n.parent
}

func (n *node) SetParent(parent Node) {
	n.parent = parent
}

func (n *node) Children() []Node {
	return n.children
}

func (n *node) SetChildren(children []Node) {
	n.children = children
}

func (n *node) HasChildren() bool {
	if len(n.children) > 0 {
		return true
	}
	return false
}

func (n *node) Raw() []byte {
	return n.raw
}

func (n *node) SetRaw(raw []byte) {
	n.raw = raw
}
