package compono

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var voidElements = map[string]struct{}{
	"area":   {},
	"base":   {},
	"br":     {},
	"col":    {},
	"embed":  {},
	"hr":     {},
	"img":    {},
	"input":  {},
	"link":   {},
	"meta":   {},
	"source": {},
	"track":  {},
	"wbr":    {},
}

// tokenizeGolden walks html.NewTokenizer over s and appends a canonical
// serialization of the element structure to b. Void elements and
// self-closing tags are written without a matching end tag; other tags are
// paired. Text, comment, and doctype tokens are ignored. Returns false if the
// tag stack is unbalanced (unclosed open tag or stray end tag).
func tokenizeGolden(b *strings.Builder, s string, caseName string, t *testing.T) bool {
	t.Helper()

	tok := html.NewTokenizer(strings.NewReader(s))
	stack := make([]string, 0, 8)
	for {
		tt := tok.Next()
		if tt == html.ErrorToken {
			if err := tok.Err(); err != nil && err != io.EOF {
				t.Errorf("%s: tokenizer error: %v", caseName, err)
			}
			break
		}
		switch tt {
		case html.StartTagToken:
			nameBytes, _ := tok.TagName()
			name := string(nameBytes)
			if _, isVoid := voidElements[name]; isVoid {
				b.WriteByte('<')
				b.WriteString(name)
				b.WriteString("/>")
				continue
			}
			b.WriteByte('<')
			b.WriteString(name)
			b.WriteByte('>')
			stack = append(stack, name)
		case html.SelfClosingTagToken:
			nameBytes, _ := tok.TagName()
			name := string(nameBytes)
			b.WriteByte('<')
			b.WriteString(name)
			b.WriteString("/>")
		case html.EndTagToken:
			nameBytes, _ := tok.TagName()
			name := string(nameBytes)
			if len(stack) == 0 || stack[len(stack)-1] != name {
				t.Errorf("%s: unexpected </%s>", caseName, name)
				return false
			}
			stack = stack[:len(stack)-1]
			b.WriteString("</")
			b.WriteString(name)
			b.WriteByte('>')
		case html.TextToken, html.CommentToken, html.DoctypeToken:
		default:
			t.Errorf("%s: unexpected token type %v", caseName, tt)
			return false
		}
	}
	if len(stack) > 0 {
		t.Errorf("%s: unclosed tags: %s", caseName, strings.Join(stack, ","))
		return false
	}
	return true
}

// serializeFragment serializes the nodes returned by html.ParseFragment into
// the same canonical form used by tokenizeGolden. Only element nodes are
// emitted; text, comment, doctype, and other node kinds are ignored. Void
// elements are emitted as <name/> without a matching end tag.
func serializeFragment(b *strings.Builder, nodes []*html.Node) {
	for _, n := range nodes {
		serializeNode(b, n)
	}
}

func serializeNode(b *strings.Builder, n *html.Node) {
	if n == nil || n.Type != html.ElementNode {
		return
	}
	if _, isVoid := voidElements[n.Data]; isVoid {
		b.WriteByte('<')
		b.WriteString(n.Data)
		b.WriteString("/>")
		return
	}
	b.WriteByte('<')
	b.WriteString(n.Data)
	b.WriteByte('>')
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		serializeNode(b, c)
	}
	b.WriteString("</")
	b.WriteString(n.Data)
	b.WriteByte('>')
}

func TestGoldenOutputsParseUnchanged(t *testing.T) {
	files, err := filepath.Glob("testdata/*/output.golden")
	require.NoError(t, err)

	for _, path := range files {
		caseName := filepath.Base(filepath.Dir(path))
		t.Run(caseName, func(t *testing.T) {
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			s := strings.TrimSpace(string(data))
			if s == "" {
				return
			}

			wantBuilder := &strings.Builder{}
			if !tokenizeGolden(wantBuilder, s, caseName, t) {
				return
			}
			want := wantBuilder.String()

			nodes, err := html.ParseFragment(strings.NewReader(s), &html.Node{
				Type:     html.ElementNode,
				Data:     "body",
				DataAtom: atom.Body,
			})
			require.NoError(t, err)

			gotBuilder := &strings.Builder{}
			serializeFragment(gotBuilder, nodes)
			got := gotBuilder.String()

			assert.Equal(t, want, got, "case %s", caseName)
		})
	}
}
