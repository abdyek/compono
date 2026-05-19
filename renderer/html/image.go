package html

import (
	"html"
	"sort"
	"strconv"
	"strings"

	"github.com/umono-cms/compono/ast"
)

type image struct {
	renderer *renderer
}

type imageVariant struct {
	url      string
	width    int
	widthRaw string
	mimeType string
}

func newImage(rend *renderer) builtinComponent {
	return &image{
		renderer: rend,
	}
}

func (img *image) New() builtinComponent {
	return newImage(img.renderer)
}

func (_ *image) Name() string {
	return "IMAGE"
}

func (img *image) Render(invoker renderableNode, node ast.Node) string {
	media := img.resolveArg(invoker, node, "media")
	alt := img.resolveArg(invoker, node, "alt")

	renderedImg := `<img src="` + html.EscapeString(img.recordField(media, "url")) + `"` +
		` alt="` + html.EscapeString(strings.TrimSpace(alt.Raw)) + `"` +
		` width="` + html.EscapeString(img.recordField(media, "width")) + `"` +
		` height="` + html.EscapeString(img.recordField(media, "height")) + `">`

	variants := img.variants(media)
	if len(variants) == 0 {
		return img.wrap(renderedImg)
	}

	grouped := map[string][]imageVariant{}
	order := []string{}
	for _, variant := range variants {
		if _, ok := grouped[variant.mimeType]; !ok {
			order = append(order, variant.mimeType)
		}
		grouped[variant.mimeType] = append(grouped[variant.mimeType], variant)
	}

	sources := make([]string, 0, len(order))
	for _, mimeType := range order {
		items := grouped[mimeType]
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].width < items[j].width
		})

		srcsetParts := make([]string, 0, len(items))
		for _, item := range items {
			srcsetParts = append(srcsetParts, html.EscapeString(item.url)+" "+item.widthRaw+"w")
		}

		sources = append(sources, `<source type="`+html.EscapeString(mimeType)+`" srcset="`+strings.Join(srcsetParts, ", ")+`">`)
	}

	return img.wrap(`<picture>` + strings.Join(sources, "") + renderedImg + `</picture>`)
}

func (_ *image) wrap(rendered string) string {
	return `<compono-image>` + rendered + `</compono-image>`
}

func (img *image) resolveArg(invoker renderableNode, compCall ast.Node, name string) ast.ResolvedValue {
	arg := ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(compCall), name)
	if arg != nil {
		return ast.ResolveCompCallArgValue(img.renderer.root, arg, getAncestorsByInvoker(invoker), compCall)
	}
	return ast.ResolveParamDefaultFromCompCall(img.renderer.root, compCall, name)
}

func (img *image) recordField(record ast.ResolvedValue, key string) string {
	field, ok := record.Fields[key]
	if !ok {
		return ""
	}
	return strings.TrimSpace(field.Raw)
}

func (img *image) variants(media ast.ResolvedValue) []imageVariant {
	variants, ok := media.Fields["variants"]
	if !ok || variants.Type != "array" {
		return nil
	}

	result := make([]imageVariant, 0, len(variants.Items))
	for _, item := range variants.Items {
		if item.Type != "record" {
			continue
		}

		widthRaw := img.recordField(item, "width")
		width, err := strconv.Atoi(widthRaw)
		if err != nil {
			continue
		}

		result = append(result, imageVariant{
			url:      img.recordField(item, "url"),
			width:    width,
			widthRaw: widthRaw,
			mimeType: img.recordField(item, "mime-type"),
		})
	}

	return result
}
