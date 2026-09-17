package rule

import "github.com/umono-cms/compono/selector"

type link struct{}

func newLink() Rule {
	return &link{}
}

func (_ *link) Name() string {
	return "link"
}

func (_ *link) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewMarkdownLink(selector.MarkdownLinkWhole),
	}
}

func (_ *link) Rules() []Rule {
	return []Rule{
		newLinkText(),
		newLinkURL(),
	}
}

type linkText struct{}

func newLinkText() Rule {
	return &linkText{}
}

func (_ *linkText) Name() string {
	return "link-text"
}

func (_ *linkText) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewMarkdownLink(selector.MarkdownLinkText),
	}
}

func (_ *linkText) Rules() []Rule {
	return []Rule{
		newStrong(),
		newEm(),
		newInlineCode(),
		newContextRef(),
		newParamRef(),
		newPlain(),
	}
}

type linkURL struct{}

func newLinkURL() Rule {
	return &linkURL{}
}

func (_ *linkURL) Name() string {
	return "link-url"
}

func (_ *linkURL) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewMarkdownLink(selector.MarkdownLinkURL),
	}
}

func (_ *linkURL) Rules() []Rule {
	return []Rule{
		newContextRef(),
		newParamRef(),
		newPlain(),
	}
}
