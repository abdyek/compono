package html

import (
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/internal/attrhook"
)

type frame = attrhook.Frame
type attributeHookFunc = attrhook.AttributeHookFunc

var attributeNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func (r *renderer) pushLocalFrame(name string, def ast.Node, signature string, call ast.Node) {
	r.frameStack = append(r.frameStack, frame{Name: name, Kind: attrhook.FrameLocal})
	r.rendering = append(r.rendering, renderingFrame{def: def, signature: signature})
	r.callNodes = append(r.callNodes, call)
}

func (r *renderer) pushGlobalFrame(name string, def ast.Node, signature string, call ast.Node) {
	r.frameStack = append(r.frameStack, frame{
		Name:      name,
		Kind:      attrhook.FrameGlobal,
		ScopePath: globalScopePath(def),
	})
	r.rendering = append(r.rendering, renderingFrame{def: def, signature: signature})
	r.callNodes = append(r.callNodes, call)
}

func (r *renderer) popFrame() {
	if len(r.frameStack) == 0 {
		return
	}
	r.frameStack = r.frameStack[:len(r.frameStack)-1]
	r.rendering = r.rendering[:len(r.rendering)-1]
	r.callNodes = r.callNodes[:len(r.callNodes)-1]
}

// isRendering reports whether def is being rendered with signature.
func (r *renderer) isRendering(def ast.Node, signature string) bool {
	for _, rf := range r.rendering {
		if rf.def == def && rf.signature == signature {
			return true
		}
	}
	return false
}

func globalScopePath(def ast.Node) []string {
	names := []string{}
	for _, anc := range ast.GetAncestors(def) {
		if !ast.IsRuleName(anc, "global-comp-def") {
			continue
		}
		nameNode := ast.FindNodeByRuleName(anc.Children(), "global-comp-name")
		if nameNode == nil {
			continue
		}
		names = append(names, strings.TrimSpace(string(nameNode.Raw())))
	}

	path := make([]string, 0, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		path = append(path, names[i])
	}
	return path
}

func (r *renderer) callAttributeHook(name string, conflicts map[string]bool) map[string]string {
	if r.attributeHook == nil {
		return nil
	}

	attrs := r.attributeHook(name, r.frameStack)

	attrNames := make([]string, 0, len(attrs))
	for attrName := range attrs {
		attrNames = append(attrNames, attrName)
	}
	sort.Strings(attrNames)

	for _, attrName := range attrNames {
		if !attributeNamePattern.MatchString(attrName) {
			r.recordAttributeError(fmt.Errorf("%w: %q returned for %s", attrhook.ErrInvalidAttributeName, attrName, name))
			continue
		}
		if conflicts[attrName] {
			r.recordAttributeError(fmt.Errorf("%w: %q returned for %s", attrhook.ErrAttributeConflict, attrName, name))
		}
	}
	return attrs
}

func (r *renderer) recordAttributeError(err error) {
	if r.attrErr == nil {
		r.attrErr = err
	}
}

func formatHookAttributes(attrs map[string]string, exclude ...string) string {
	if len(attrs) == 0 {
		return ""
	}

	excluded := make(map[string]bool, len(exclude))
	for _, name := range exclude {
		excluded[name] = true
	}

	names := make([]string, 0, len(attrs))
	for name := range attrs {
		if excluded[name] {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		b.WriteString(" ")
		b.WriteString(name)
		b.WriteString(`="`)
		b.WriteString(html.EscapeString(attrs[name]))
		b.WriteString(`"`)
	}
	return b.String()
}
