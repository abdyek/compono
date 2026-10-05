package errwrap

import (
	"strconv"

	"github.com/umono-cms/compono/ast"
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
