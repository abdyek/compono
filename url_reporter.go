package compono

import (
	"github.com/umono-cms/compono/internal/urlreport"
)

// URLReport is a URL that Compono wrote to the output and that the browser
// loads as a resource.
type URLReport = urlreport.URLReport

// URLReporterFunc receives every reported URL.
type URLReporterFunc = urlreport.URLReporterFunc

type urlReporterOption struct {
	fn URLReporterFunc
}

// WithURLReporter returns a ConvertOption that registers a function invoked
// once for each src or srcset URL of IMAGE in output order. It may be used at
// most once per conversion and is not permitted inside WithGlobalComponent.
// A nil function is treated as not setting the reporter.
func WithURLReporter(fn URLReporterFunc) ConvertOption {
	return &urlReporterOption{fn: fn}
}

func (o *urlReporterOption) applyConvert(_ *compono, cfg *convertConfig) error {
	if cfg.urlReporterSet {
		return NewComponoError(ErrURLReporterAlreadySet, "url reporter is already set: WithURLReporter can be used at most once per conversion")
	}
	cfg.urlReporterSet = true
	cfg.urlReporter = o.fn
	return nil
}
