package builtin

import (
	"fmt"
	"strconv"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
)

func BuildASTNodes(parent ast.Node) []ast.Node {
	builtinComps := BuiltinComponents()
	if len(builtinComps) == 0 {
		return nil
	}

	children := make([]ast.Node, 0, len(builtinComps))

	for _, comp := range builtinComps {
		builtinComp := ast.DefaultEmptyNode()
		builtinComp.SetRule(rule.NewDynamic("builtin-comp"))
		builtinComp.SetParent(parent)

		builtinCompName := ast.DefaultEmptyNode()
		builtinCompName.SetRule(rule.NewDynamic("builtin-comp-name"))
		builtinCompName.SetParent(builtinComp)
		builtinCompName.SetRaw([]byte(comp.Name))

		compParams := ast.DefaultEmptyNode()
		compParams.SetRule(rule.NewDynamic("comp-params"))
		compParams.SetParent(builtinComp)
		compParams.SetChildren(makeBuiltinCompParams(compParams, comp.Params))

		builtinComp.SetChildren([]ast.Node{builtinCompName, compParams})
		children = append(children, builtinComp)
	}

	return children
}

func makeBuiltinCompParams(parent ast.Node, params []Param) []ast.Node {
	if len(params) == 0 {
		return nil
	}

	children := make([]ast.Node, 0, len(params))
	for _, param := range params {
		compParam := ast.DefaultEmptyNode()
		compParam.SetRule(rule.NewDynamic("comp-param"))
		compParam.SetParent(parent)

		compParamName := ast.DefaultEmptyNode()
		compParamName.SetRule(rule.NewDynamic("comp-param-name"))
		compParamName.SetParent(compParam)
		compParamName.SetRaw([]byte(param.Name))

		compParamType := ast.DefaultEmptyNode()
		compParamType.SetRule(rule.NewDynamic("comp-param-type"))
		compParamType.SetParent(compParam)

		typedParam := ast.DefaultEmptyNode()
		typedParam.SetRule(rule.NewDynamic(paramTypeRuleName(param.Type)))
		typedParam.SetParent(compParamType)

		compParamDefaValue := ast.DefaultEmptyNode()
		compParamDefaValue.SetRule(rule.NewDynamic("comp-param-defa-value"))
		compParamDefaValue.SetParent(typedParam)
		compParamDefaValue.SetRaw([]byte(paramDefaultValue(param)))

		typedParam.SetChildren([]ast.Node{compParamDefaValue})
		compParamType.SetChildren([]ast.Node{typedParam})
		compParam.SetChildren([]ast.Node{compParamName, compParamType})

		children = append(children, compParam)
	}

	return children
}

func paramTypeRuleName(paramType ParamType) string {
	switch paramType {
	case StringType:
		return "comp-string-param"
	case NumberType:
		return "comp-number-param"
	case BoolType:
		return "comp-bool-param"
	case ComponentType:
		return "comp-comp-param"
	default:
		return "comp-string-param"
	}
}

func paramDefaultValue(param Param) string {
	switch param.Type {
	case StringType:
		if v, ok := param.DefaultValue.(string); ok {
			return v
		}
	case NumberType:
		switch v := param.DefaultValue.(type) {
		case int:
			return strconv.Itoa(v)
		case int8:
			return strconv.FormatInt(int64(v), 10)
		case int16:
			return strconv.FormatInt(int64(v), 10)
		case int32:
			return strconv.FormatInt(int64(v), 10)
		case int64:
			return strconv.FormatInt(v, 10)
		case uint:
			return strconv.FormatUint(uint64(v), 10)
		case uint8:
			return strconv.FormatUint(uint64(v), 10)
		case uint16:
			return strconv.FormatUint(uint64(v), 10)
		case uint32:
			return strconv.FormatUint(uint64(v), 10)
		case uint64:
			return strconv.FormatUint(v, 10)
		case float32:
			return strconv.FormatFloat(float64(v), 'f', -1, 32)
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		}
	case BoolType:
		if v, ok := param.DefaultValue.(bool); ok {
			return strconv.FormatBool(v)
		}
	case ComponentType:
		if v, ok := param.DefaultValue.(string); ok {
			return v
		}
	}

	return fmt.Sprint(param.DefaultValue)
}
