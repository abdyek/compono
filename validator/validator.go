package validator

import (
	"errors"
	"fmt"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

type Validator interface {
	Validate(ast.Node) error
}

func DefaultValidator() Validator {
	return &validator{
		requiredChildren: map[string][]string{
			"root": []string{
				"root-content",
				"global-comp-def-wrapper",
			},
			"local-comp-def": []string{
				"local-comp-def-head",
				"local-comp-def-content",
			},
			"global-comp-def": []string{
				"global-comp-def-content",
			},
			"block-comp-call": []string{
				"comp-call-name",
			},
			"inline-comp-call": []string{
				"comp-call-name",
			},
			"param-ref": []string{
				"param-ref-name",
			},
		},
	}
}

type validator struct {
	requiredChildren map[string][]string
}

func (v *validator) Validate(root ast.Node) error {
	return v.validateNode(root)
}

func (v *validator) validateNode(node ast.Node) error {
	if node.Rule() == nil {
		return errors.New("Each node must have a rule.")
	}
	if err := v.checkChildren(node); err != nil {
		return err
	}
	for _, child := range node.Children() {
		if err := v.validateNode(child); err != nil {
			return err
		}
	}
	return nil
}

func (v *validator) checkChildren(node ast.Node) error {
	ruleName := node.Rule().Name()
	requiredChildren, ok := v.requiredChildren[ruleName]
	if !ok {
		return nil
	}

	childrenNames := []string{}
	for _, child := range node.Children() {
		if child.Rule() == nil {
			return errors.New("Each node must have a rule.")
		}
		childrenNames = append(childrenNames, child.Rule().Name())
	}

	for _, rName := range requiredChildren {
		if !util.InSliceString(rName, childrenNames) {
			return fmt.Errorf("%s must have %s", ruleName, rName)
		}
	}

	return nil
}
