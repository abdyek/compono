package errwrap

import (
	"strconv"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/util"
)

var imageSupportedMimeTypes = []string{
	"image/jpeg",
	"image/png",
	"image/webp",
	"image/gif",
	"image/avif",
}

type imageError struct {
	title   string
	message string
}

type imageComponentTarget struct {
	name  string
	scope ast.Node
}

func wrongImageArgType() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isImageBuiltinComponent(),
			hasWrongTypeArgs(),
		},
		title:   staticTitle("Wrong argument type"),
		message: wrongArgTypeMsg,
		block:   blockFromRuleName,
	}
}

func invalidImage() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			not(isInsideCompDef()),
			isImageWithSpecificError(),
		},
		title: func(ctx *wrapContext, node ast.Node) string {
			return getImageError(ctx, node).title
		},
		message: func(ctx *wrapContext, node ast.Node) string {
			return getImageError(ctx, node).message
		},
		block: blockFromRuleName,
	}
}

func isImageBuiltinComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		if getCompCallNameStr(node) != "IMAGE" {
			return false
		}

		compDef := findCompDef(ctx.root, node, "IMAGE")
		return compDef != nil && ast.IsRuleName(compDef, "builtin-comp")
	}
}

func isImageWithSpecificError() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return getImageError(ctx, node).title != ""
	}
}

func getImageError(ctx *wrapContext, node ast.Node) imageError {
	if !ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"}) {
		return imageError{}
	}

	return walkImageCallTree(ctx, node, func(target ast.Node, invokerAncestors []ast.Node) imageError {
		return getImageErrorForCompCalls(ctx, target, invokerAncestors)
	})
}

func walkImageCallTree(ctx *wrapContext, ownerCompCall ast.Node, visit func(ast.Node, []ast.Node) imageError) imageError {
	seen := map[ast.Node]bool{}

	var walk func(ast.Node, []ast.Node) imageError
	walk = func(current ast.Node, invokerAncestors []ast.Node) imageError {
		if seen[current] {
			return imageError{}
		}
		seen[current] = true

		if getCompCallNameStr(current) == "IMAGE" {
			if err := visit(current, invokerAncestors); err.title != "" {
				return err
			}
		}

		compName := getCompCallNameStr(current)
		if compName == "" {
			return imageError{}
		}

		compDef := findCompDef(ctx.root, current, compName)
		if compDef == nil {
			return imageError{}
		}

		compDefContent := getCompDefContent(compDef)
		if compDefContent == nil {
			return imageError{}
		}

		for _, nested := range ast.FilterNodesInTree(compDefContent, func(child ast.Node) bool {
			return ast.IsRuleNameOneOf(child, []string{"block-comp-call", "inline-comp-call"})
		}) {
			if err := walk(nested, append([]ast.Node{current}, invokerAncestors...)); err.title != "" {
				return err
			}
		}

		return imageError{}
	}

	return walk(ownerCompCall, ast.GetAncestors(ownerCompCall))
}

func getImageErrorForCompCalls(ctx *wrapContext, targetCompCall ast.Node, invokerAncestors []ast.Node) imageError {
	if getCompCallNameStr(targetCompCall) != "IMAGE" {
		return imageError{}
	}

	targetCompDef := findCompDef(ctx.root, targetCompCall, "IMAGE")
	if targetCompDef == nil || !ast.IsRuleName(targetCompDef, "builtin-comp") {
		return imageError{}
	}

	media := resolveImageArg(ctx, targetCompCall, invokerAncestors, "media")
	if key := resolvedValueMissingContextKey(media); key != "" {
		return imageError{
			title:   "Unknown key",
			message: "The key **" + key + "** is not injected.",
		}
	}
	if key := resolvedValueMissingContextKey(resolveImageArg(ctx, targetCompCall, invokerAncestors, "alt")); key != "" {
		return imageError{
			title:   "Unknown key",
			message: "The key **" + key + "** is not injected.",
		}
	}

	if err := getImageUnsupportedMimeTypeError(media); err.title != "" {
		return err
	}
	if err := getImageInvalidDimensionError(media); err.title != "" {
		return err
	}
	if err := getImageDuplicateVariantError(media); err.title != "" {
		return err
	}
	if err := getImageInconsistentAspectRatioError(media); err.title != "" {
		return err
	}

	return imageError{}
}

func resolveImageArg(ctx *wrapContext, targetCompCall ast.Node, invokerAncestors []ast.Node, name string) ast.ResolvedValue {
	arg := ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(targetCompCall), name)
	if arg != nil {
		return ast.ResolveCompCallArgValue(ctx.root, arg, invokerAncestors, targetCompCall)
	}

	return ast.ResolveParamDefaultFromCompCall(ctx.root, targetCompCall, name)
}

func getImageUnsupportedMimeTypeError(media ast.ResolvedValue) imageError {
	mimeType := imageRecordStringField(media, "mime-type")
	if mimeType != "" && !util.InSliceString(mimeType, imageSupportedMimeTypes) {
		return imageError{
			title:   "Unsupported mime-type",
			message: "The mime-type **" + mimeType + "** is unsupported.",
		}
	}

	for _, variant := range imageVariants(media) {
		mimeType = imageRecordStringField(variant, "mime-type")
		if mimeType != "" && !util.InSliceString(mimeType, imageSupportedMimeTypes) {
			return imageError{
				title:   "Unsupported mime-type",
				message: "The mime-type **" + mimeType + "** is unsupported.",
			}
		}
	}

	return imageError{}
}

func getImageInvalidDimensionError(media ast.ResolvedValue) imageError {
	for _, field := range []string{"width", "height"} {
		value, ok := imageRecordIntField(media, field)
		if ok && value <= 0 {
			return imageError{
				title:   "Invalid dimension",
				message: "The value of **" + field + "** must be greater than 0.",
			}
		}
	}

	for _, variant := range imageVariants(media) {
		for _, field := range []string{"width", "height"} {
			value, ok := imageRecordIntField(variant, field)
			if ok && value <= 0 {
				return imageError{
					title:   "Invalid dimension",
					message: "The value of **" + field + "** must be greater than 0.",
				}
			}
		}
	}

	return imageError{}
}

func getImageDuplicateVariantError(media ast.ResolvedValue) imageError {
	seen := map[string]string{}

	for _, variant := range imageVariants(media) {
		mimeType := imageRecordStringField(variant, "mime-type")
		width := imageRecordStringField(variant, "width")
		key := mimeType + "\x00" + width

		if _, ok := seen[key]; ok {
			return imageError{
				title:   "Duplicate variant",
				message: "The variant with mime-type **" + mimeType + "** and width **" + width + "** is defined more than once.",
			}
		}

		seen[key] = width
	}

	return imageError{}
}

func getImageInconsistentAspectRatioError(media ast.ResolvedValue) imageError {
	mediaWidth, ok := imageRecordIntField(media, "width")
	if !ok {
		return imageError{}
	}
	mediaHeight, ok := imageRecordIntField(media, "height")
	if !ok {
		return imageError{}
	}

	for _, variant := range imageVariants(media) {
		variantWidth, ok := imageRecordIntField(variant, "width")
		if !ok {
			continue
		}
		variantHeight, ok := imageRecordIntField(variant, "height")
		if !ok {
			continue
		}

		if !imagePreservesAspectRatio(mediaWidth, mediaHeight, variantWidth, variantHeight) {
			return imageError{
				title:   "Inconsistent aspect ratio",
				message: "All variants must preserve the aspect ratio of the main media.",
			}
		}
	}

	return imageError{}
}

func imagePreservesAspectRatio(mediaWidth, mediaHeight, variantWidth, variantHeight int) bool {
	expectedHeightNumerator := variantWidth * mediaHeight
	expectedHeightFloor := expectedHeightNumerator / mediaWidth
	expectedHeightCeil := divideAndCeil(expectedHeightNumerator, mediaWidth)

	if variantHeight == expectedHeightFloor || variantHeight == expectedHeightCeil {
		return true
	}

	expectedWidthNumerator := variantHeight * mediaWidth
	expectedWidthFloor := expectedWidthNumerator / mediaHeight
	expectedWidthCeil := divideAndCeil(expectedWidthNumerator, mediaHeight)

	return variantWidth == expectedWidthFloor || variantWidth == expectedWidthCeil
}

func divideAndCeil(numerator, denominator int) int {
	return (numerator + denominator - 1) / denominator
}

func imageVariants(media ast.ResolvedValue) []ast.ResolvedValue {
	variants, ok := media.Fields["variants"]
	if !ok || variants.Type != "array" {
		return nil
	}
	return variants.Items
}

func imageRecordStringField(record ast.ResolvedValue, key string) string {
	field, ok := record.Fields[key]
	if !ok {
		return ""
	}
	return field.Raw
}

func imageRecordIntField(record ast.ResolvedValue, key string) (int, bool) {
	raw := imageRecordStringField(record, key)
	if raw == "" {
		return 0, false
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}

	return value, true
}

func imageErrorForComponentTarget(ctx *wrapContext, caller ast.Node, target imageComponentTarget, parentInvokers []ast.Node, seen map[string]bool) imageError {
	if target.name == "" {
		return imageError{}
	}

	signature := target.name
	if target.scope != nil {
		signature += "\x00" + target.scope.Rule().Name() + "\x00" + string(target.scope.Raw())
	}
	if seen[signature] {
		return imageError{}
	}
	seen[signature] = true
	defer delete(seen, signature)

	syntheticCall := createSyntheticCompCall(caller, target.name)
	invokerAncestors := imageComponentInvokerAncestors(caller, syntheticCall, parentInvokers)

	if target.name == "IMAGE" {
		if len(getBuiltinSchemaMismatchArgNamesForCompCall(ctx, syntheticCall, syntheticCall)) > 0 {
			return imageError{
				title:   "Invalid built-in arguments",
				message: invalidBuiltinCompCallSchemaMsg(ctx, syntheticCall),
			}
		}
		return getImageErrorForCompCalls(ctx, syntheticCall, invokerAncestors)
	}

	compDef := findWebGridItemComponentDef(ctx.root, caller, target.name, target.scope)
	if compDef == nil || ast.IsRuleName(compDef, "builtin-comp") {
		return imageError{}
	}

	content := getCompDefContent(compDef)
	if content == nil {
		return imageError{}
	}

	for _, nested := range ast.FilterNodesInTree(content, func(node ast.Node) bool {
		if ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"}) {
			return true
		}
		return ast.IsRuleName(node, "param-ref") && hasCompCallArgsNode(node)
	}) {
		if ast.IsRuleName(nested, "param-ref") {
			nextTarget := resolveParamRefComponentTarget(ctx, nested, invokerAncestors)
			if nextTarget.name == "" {
				continue
			}
			if err := imageErrorForComponentTarget(ctx, nested, nextTarget, append([]ast.Node{nested}, invokerAncestors...), seen); err.title != "" {
				return err
			}
			continue
		}

		nestedName := getCompCallNameStr(nested)
		if nestedName == "" {
			continue
		}
		if nestedName == "IMAGE" {
			if err := getImageErrorForCompCalls(ctx, nested, invokerAncestors); err.title != "" {
				return err
			}
			continue
		}
		if err := imageErrorForComponentTarget(ctx, nested, imageComponentTarget{
			name:  nestedName,
			scope: ast.GetLocalCompSourceFromNode(nested, ctx.root),
		}, append([]ast.Node{nested}, invokerAncestors...), seen); err.title != "" {
			return err
		}
	}

	return imageError{}
}

func imageComponentInvokerAncestors(caller ast.Node, syntheticCall ast.Node, parentInvokers []ast.Node) []ast.Node {
	if len(parentInvokers) == 0 || parentInvokers[0] != caller {
		return append([]ast.Node{syntheticCall}, parentInvokers...)
	}

	if ast.IsRuleName(caller, "param-ref") {
		return append([]ast.Node{caller, syntheticCall}, parentInvokers[1:]...)
	}

	return parentInvokers
}

func resolveParamRefComponentTarget(ctx *wrapContext, paramRef ast.Node, invokerAncestors []ast.Node) imageComponentTarget {
	paramName := getParamRefNameStr(paramRef)
	if paramName == "" {
		return imageComponentTarget{}
	}

	resolved := ast.ResolveParamFromAncestors(ctx.root, paramName, ast.GetParamRefAccessors(paramRef), invokerAncestors)
	if resolved.IsZero() {
		if compDef := findEnclosingCompDef(paramRef); compDef != nil {
			resolved = ast.ApplyAccessors(ast.ResolveCompParamDefaultFromCompDef(ctx.root, compDef, paramName), ast.GetParamRefAccessors(paramRef))
		}
	}
	if resolved.Type != "comp" || resolved.Raw == "" {
		return imageComponentTarget{}
	}

	return imageComponentTarget{
		name:  resolved.Raw,
		scope: resolved.Scope,
	}
}

func createSyntheticCompCall(parent ast.Node, name string) ast.Node {
	compCall := ast.DefaultEmptyNode()
	compCall.SetRule(rule.NewDynamic("block-comp-call"))
	compCall.SetParent(parent)

	compCallName := ast.DefaultEmptyNode()
	compCallName.SetRule(rule.NewDynamic("comp-call-name"))
	compCallName.SetParent(compCall)
	compCallName.SetRaw([]byte(name))

	compCall.SetChildren([]ast.Node{compCallName})
	return compCall
}
