package builtin

import (
	"fmt"

	"github.com/umono-cms/compono/ast"
)

type ValueKind int

const (
	StringKind ValueKind = iota + 1
	IntegerKind
	BoolKind
	ComponentKind
	ArrayKind
	RecordKind
)

func (k ValueKind) String() string {
	switch k {
	case StringKind:
		return "string"
	case IntegerKind:
		return "integer"
	case BoolKind:
		return "bool"
	case ComponentKind:
		return "component"
	case ArrayKind:
		return "array"
	case RecordKind:
		return "record"
	default:
		return "unknown"
	}
}

type ValueSchema interface {
	Kind() ValueKind
}

type ScalarSchema struct {
	valueKind     ValueKind
	AllowedValues []any
	Matcher       func(ast.ResolvedValue) bool
}

func (s *ScalarSchema) Kind() ValueKind {
	return s.valueKind
}

func (s *ScalarSchema) OneOf(values ...any) *ScalarSchema {
	s.AllowedValues = append([]any(nil), values...)
	return s
}

func (s *ScalarSchema) Matching(fn func(ast.ResolvedValue) bool) *ScalarSchema {
	s.Matcher = fn
	return s
}

type ArrayElementMode int

const (
	ArrayElementsHomogeneous ArrayElementMode = iota + 1
	ArrayElementsAnyOf
	ArrayElementsTuple
)

type ArraySchema struct {
	ElementSchemas []ValueSchema
	ElementsMode   ArrayElementMode
	MinLen         int
	MaxLen         int
}

func (s *ArraySchema) Kind() ValueKind {
	return ArrayKind
}

func (s *ArraySchema) Min(length int) *ArraySchema {
	s.MinLen = length
	return s
}

func (s *ArraySchema) Max(length int) *ArraySchema {
	s.MaxLen = length
	return s
}

type RecordSchema struct {
	Fields           []RecordField
	AllowUnknownKeys bool
	MinFields        int
	MaxFields        int
}

func (s *RecordSchema) Kind() ValueKind {
	return RecordKind
}

func (s *RecordSchema) AllowUnknown() *RecordSchema {
	s.AllowUnknownKeys = true
	return s
}

func (s *RecordSchema) DisallowUnknownKeys() *RecordSchema {
	s.AllowUnknownKeys = false
	return s
}

func (s *RecordSchema) Min(length int) *RecordSchema {
	s.MinFields = length
	return s
}

func (s *RecordSchema) Max(length int) *RecordSchema {
	s.MaxFields = length
	return s
}

type RecordField struct {
	Name        string
	Schema      ValueSchema
	IsRequired  bool
	Description string
}

func (f RecordField) RequiredField() RecordField {
	f.IsRequired = true
	return f
}

func (f RecordField) Required() RecordField {
	return f.RequiredField()
}

func (f RecordField) Optional() RecordField {
	f.IsRequired = false
	return f
}

func (f RecordField) WithDescription(description string) RecordField {
	f.Description = description
	return f
}

type Definition struct {
	Name             string
	Params           []Param
	InlineRenderable bool
}

type Param struct {
	Name         string
	Schema       ValueSchema
	DefaultValue any
	Description  string
}

func String() *ScalarSchema {
	return &ScalarSchema{valueKind: StringKind}
}

func Integer() *ScalarSchema {
	return &ScalarSchema{valueKind: IntegerKind}
}

func Bool() *ScalarSchema {
	return &ScalarSchema{valueKind: BoolKind}
}

func ComponentValue() *ScalarSchema {
	return &ScalarSchema{valueKind: ComponentKind}
}

func Component() *ScalarSchema {
	return ComponentValue()
}

func Enum(values ...any) *ScalarSchema {
	if len(values) == 0 {
		panic("builtin.Enum requires at least one value")
	}

	kind, ok := scalarKindOf(values[0])
	if !ok {
		panic(fmt.Sprintf("builtin.Enum does not support %T", values[0]))
	}

	for _, value := range values[1:] {
		currentKind, currentOK := scalarKindOf(value)
		if !currentOK {
			panic(fmt.Sprintf("builtin.Enum does not support %T", value))
		}
		if currentKind != kind {
			panic("builtin.Enum requires values of the same scalar kind")
		}
	}

	return (&ScalarSchema{valueKind: kind}).OneOf(values...)
}

func ArrayOf(schema ValueSchema) *ArraySchema {
	return &ArraySchema{
		ElementSchemas: []ValueSchema{schema},
		ElementsMode:   ArrayElementsHomogeneous,
		MinLen:         -1,
		MaxLen:         -1,
	}
}

func Array(schemas ...ValueSchema) *ArraySchema {
	return &ArraySchema{
		ElementSchemas: append([]ValueSchema(nil), schemas...),
		ElementsMode:   ArrayElementsAnyOf,
		MinLen:         -1,
		MaxLen:         -1,
	}
}

func Tuple(schemas ...ValueSchema) *ArraySchema {
	return &ArraySchema{
		ElementSchemas: append([]ValueSchema(nil), schemas...),
		ElementsMode:   ArrayElementsTuple,
		MinLen:         -1,
		MaxLen:         -1,
	}
}

func Record(fields ...RecordField) *RecordSchema {
	return &RecordSchema{
		Fields:           append([]RecordField(nil), fields...),
		AllowUnknownKeys: true,
		MinFields:        -1,
		MaxFields:        -1,
	}
}

func Field(name string, schema ValueSchema) RecordField {
	return RecordField{
		Name:   name,
		Schema: schema,
	}
}
