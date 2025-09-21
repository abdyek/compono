package ast

import "github.com/umono-cms/compono/rule"

type Node interface {
	Rule() rule.Rule
	SetRule(rule.Rule)
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
	rule     rule.Rule
	parent   Node
	children []Node
	raw      []byte
}

func (n *node) Rule() rule.Rule {
	return n.rule
}

func (n *node) SetRule(rule rule.Rule) {
	n.rule = rule
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
