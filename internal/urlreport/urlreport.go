// Package urlreport defines the types that Compono uses to report URLs that
// the browser will load as resources.
package urlreport

// URLReport is a URL that Compono wrote to the output and that the browser
// loads as a resource.
type URLReport struct {
	Element   string
	Attribute string
	URL       string
}

// URLReporterFunc receives every reported URL.
type URLReporterFunc func(r URLReport)
