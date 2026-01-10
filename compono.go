package compono

import (
	"fmt"
	"io"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/errwrap"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/parser"
	"github.com/umono-cms/compono/renderer"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/util"
	"github.com/umono-cms/compono/validator"
)

type ErrorCode int

const (
	ErrInvalidGlobalName ErrorCode = iota + 1
	ErrGlobalAlreadyRegistered
	ErrGlobalNotExist
	ErrInvalidAST
	ErrRender
)

type Compono interface {
	Convert(source []byte, writer io.Writer) error
	ConvertGlobalComponent(string, []byte, io.Writer) error
	RegisterGlobalComponent(string, []byte) error
	UnregisterGlobalComponent(string) error
	Parser() parser.Parser
	SetParser(parser.Parser)
	Renderer() renderer.Renderer
	SetRenderer(renderer.Renderer)
	Validator() validator.Validator
	SetValidator(validator.Validator)
	ErrorWrapper() errwrap.ErrorWrapper
	SetErrorWrapper(errwrap.ErrorWrapper)
	Logger() logger.Logger
	SetLogger(logger.Logger)
}

func New() Compono {
	log := logger.NewLogger()

	p := parser.DefaultParser(log)
	r := renderer.DefaultRenderer(log)
	v := validator.DefaultValidator()
	ew := errwrap.DefaultErrorWrapper()

	gw := ast.DefaultEmptyNode()
	gw.SetRule(rule.NewGlobalCompDefWrapper())

	return &compono{
		parser:        p,
		renderer:      r,
		validator:     v,
		errorWrapper:  ew,
		logger:        log,
		globalWrapper: gw,
	}
}

type compono struct {
	parser        parser.Parser
	renderer      renderer.Renderer
	validator     validator.Validator
	errorWrapper  errwrap.ErrorWrapper
	logger        logger.Logger
	globalWrapper ast.Node
}

func (c *compono) Convert(source []byte, writer io.Writer) error {
	if len(source) == 0 {
		return nil
	}

	root := c.parser.Parse(source, ast.DefaultRootNode())

	c.globalWrapper.SetParent(root)
	root.SetChildren(append(root.Children(), c.globalWrapper))

	err := c.validator.Validate(root)
	if err != nil {
		return NewComponoError(ErrInvalidAST, err.Error())
	}

	c.errorWrapper.Wrap(root)

	err = c.renderer.Render(writer, root)
	if err != nil {
		return NewComponoError(ErrRender, err.Error())
	}
	return nil
}

func (c *compono) ConvertGlobalComponent(name string, source []byte, writer io.Writer) error {
	// TODO: complete it
	return nil
}

func (c *compono) RegisterGlobalComponent(name string, source []byte) error {

	if !util.IsScreamingSnakeCase(name) {
		return NewComponoError(ErrInvalidGlobalName, fmt.Sprintf("invalid global component name %q: must be SCREAMING_SNAKE_CASE (digits allowed)", name))
	}

	if registered := c.getGlobalCompDefByName(name); registered != nil {
		return NewComponoError(ErrGlobalAlreadyRegistered, fmt.Sprintf("cannot register global component %q: already registered", name))
	}

	node := ast.DefaultEmptyNode()
	node.SetRule(rule.NewGlobalCompDef())

	parsed := c.parser.Parse(source, node)

	globalCompName := ast.DefaultEmptyNode()
	globalCompName.SetRule(rule.NewGlobalCompName())
	globalCompName.SetParent(parsed)
	globalCompName.SetRaw([]byte(name))

	parsed.SetChildren(append([]ast.Node{globalCompName}, parsed.Children()...))
	c.globalWrapper.SetChildren(append([]ast.Node{parsed}, c.globalWrapper.Children()...))

	return nil
}

func (c *compono) UnregisterGlobalComponent(name string) error {
	if registered := c.getGlobalCompDefByName(name); registered == nil {
		return NewComponoError(ErrGlobalNotExist, fmt.Sprintf("cannot unregister global component %q: does not exist", name))
	}

	globalComps := []ast.Node{}
	for _, gcd := range c.globalWrapper.Children() {
		if gcd.Rule().Name() != "global-comp-def" {
			continue
		}
		for _, child := range gcd.Children() {
			if !(child.Rule().Name() == "global-comp-name" && name == string(child.Raw())) {
				globalComps = append(globalComps, gcd)
				continue
			}
		}
	}

	c.globalWrapper.SetChildren(globalComps)
	return nil
}

func (c *compono) Parser() parser.Parser {
	return c.parser
}

func (c *compono) SetParser(parser parser.Parser) {
	c.parser = parser
}

func (c *compono) Renderer() renderer.Renderer {
	return c.renderer
}

func (c *compono) SetRenderer(renderer renderer.Renderer) {
	c.renderer = renderer
}

func (c *compono) Validator() validator.Validator {
	return c.validator
}

func (c *compono) SetValidator(vldtr validator.Validator) {
	c.validator = vldtr
}

func (c *compono) ErrorWrapper() errwrap.ErrorWrapper {
	return c.errorWrapper
}

func (c *compono) SetErrorWrapper(ew errwrap.ErrorWrapper) {
	c.errorWrapper = ew
}

func (c *compono) Logger() logger.Logger {
	return c.logger
}

func (c *compono) SetLogger(logger logger.Logger) {
	c.logger = logger
}

func (c *compono) getGlobalCompDefByName(name string) ast.Node {
	for _, gcd := range c.globalWrapper.Children() {
		if gcd.Rule().Name() != "global-comp-def" {
			continue
		}
		for _, child := range gcd.Children() {
			if child.Rule().Name() == "global-comp-name" && name == string(child.Raw()) {
				return gcd
			}
		}
	}
	return nil
}

type ComponoError struct {
	Code    ErrorCode
	Message string
}

func (e *ComponoError) Error() string { return e.Message }

func NewComponoError(code ErrorCode, msg string) *ComponoError {
	return &ComponoError{Code: code, Message: msg}
}
