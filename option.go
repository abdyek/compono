package compono

import "github.com/umono-cms/compono/ast"

type ConvertOption interface {
	applyConvert(*compono, *convertConfig) error
}

type convertOptionFunc func(*compono, *convertConfig) error

func (f convertOptionFunc) applyConvert(c *compono, cfg *convertConfig) error {
	return f(c, cfg)
}

type convertConfig struct {
	globalComponents []ast.Node
}

func WithGlobalComponent(name string, source []byte) ConvertOption {
	return convertOptionFunc(func(c *compono, cfg *convertConfig) error {
		globalComponent, err := c.newGlobalComponentNode(name, source)
		if err != nil {
			return err
		}

		cfg.globalComponents = append(cfg.globalComponents, globalComponent)
		return nil
	})
}
