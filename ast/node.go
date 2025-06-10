package ast

type Node interface {
	Parent() Node
	Children() []Node
	HasChildren() bool
}
