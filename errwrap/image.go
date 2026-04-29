package errwrap

import (
	"strconv"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/builtin"
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

	return walkImageCallTree(ctx, node, func(target ast.Node) imageError {
		return getImageErrorForCompCalls(ctx, node, target)
	})
}

func walkImageCallTree(ctx *wrapContext, ownerCompCall ast.Node, visit func(ast.Node) imageError) imageError {
	seen := map[ast.Node]bool{}

	var walk func(ast.Node) imageError
	walk = func(current ast.Node) imageError {
		if seen[current] {
			return imageError{}
		}
		seen[current] = true

		if getCompCallNameStr(current) == "IMAGE" {
			if err := visit(current); err.title != "" {
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
			if err := walk(nested); err.title != "" {
				return err
			}
		}

		return imageError{}
	}

	_ = ownerCompCall
	return walk(ownerCompCall)
}

func getImageErrorForCompCalls(ctx *wrapContext, ownerCompCall ast.Node, targetCompCall ast.Node) imageError {
	if getCompCallNameStr(targetCompCall) != "IMAGE" {
		return imageError{}
	}

	targetCompDef := findCompDef(ctx.root, targetCompCall, "IMAGE")
	if targetCompDef == nil || !ast.IsRuleName(targetCompDef, "builtin-comp") {
		return imageError{}
	}

	media := resolveImageArg(ctx, ownerCompCall, targetCompCall, "media")
	if !imageMediaMatchesSchema(media) {
		return imageError{
			title:   "Invalid built-in arguments",
			message: "The parameter **media** does not match the schema of the built-in component **IMAGE**.",
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

func resolveImageArg(ctx *wrapContext, ownerCompCall ast.Node, targetCompCall ast.Node, name string) ast.ResolvedValue {
	arg := ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(targetCompCall), name)
	if arg != nil {
		invokerAncestors := ast.GetAncestors(ownerCompCall)
		currentCompCall := ownerCompCall
		if ownerCompCall != targetCompCall {
			invokerAncestors = append([]ast.Node{targetCompCall, ownerCompCall}, invokerAncestors...)
			currentCompCall = targetCompCall
		}
		return ast.ResolveCompCallArgValue(ctx.root, arg, invokerAncestors, currentCompCall)
	}

	return ast.ResolveParamDefaultFromCompCall(ctx.root, targetCompCall, name)
}

func imageMediaMatchesSchema(media ast.ResolvedValue) bool {
	definition, ok := builtin.FindDefinition("IMAGE")
	if !ok {
		return false
	}

	for _, param := range definition.Params {
		if param.Name != "media" {
			continue
		}
		return builtin.MatchesResolvedValue(param.Schema, media)
	}

	return false
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

		if mediaWidth*variantHeight != variantWidth*mediaHeight {
			return imageError{
				title:   "Inconsistent aspect ratio",
				message: "All variants must preserve the aspect ratio of the main media.",
			}
		}
	}

	return imageError{}
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
