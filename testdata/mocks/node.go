package mocks

import (
	"github.com/umono-cms/compono/ast"
	rulepkg "github.com/umono-cms/compono/rule"
)

type node struct {
	rule     rulepkg.Rule
	parent   ast.Node
	children []ast.Node
	raw      []byte
}

func NewNode(rule rulepkg.Rule, parent ast.Node, children []ast.Node, raw []byte) ast.Node {
	return &node{
		rule:     rule,
		parent:   parent,
		children: children,
		raw:      raw,
	}
}

func (n *node) Rule() rulepkg.Rule {
	return n.rule
}

func (n *node) SetRule(rule rulepkg.Rule) {
	n.rule = rule
}

func (n *node) Parent() ast.Node {
	return n.parent
}

func (n *node) SetParent(parent ast.Node) {
	n.parent = parent
}

func (n *node) Children() []ast.Node {
	return n.children
}

func (n *node) SetChildren(children []ast.Node) {
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
