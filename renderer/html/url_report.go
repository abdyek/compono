package html

import (
	"regexp"
	"strconv"

	"github.com/umono-cms/compono/internal/urlreport"
)

var urlReportMarkerPattern = regexp.MustCompile("\x00compono-url-([0-9]+)\x00")

func (r *renderer) recordURL(element, attribute, url string) string {
	r.urlReports = append(r.urlReports, urlreport.URLReport{
		Element:   element,
		Attribute: attribute,
		URL:       url,
	})
	return "\x00compono-url-" + strconv.Itoa(len(r.urlReports)-1) + "\x00"
}

func (r *renderer) takeURLReports(out string) (string, []urlreport.URLReport) {
	var reports []urlreport.URLReport

	result := urlReportMarkerPattern.ReplaceAllStringFunc(out, func(match string) string {
		submatches := urlReportMarkerPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		idx, err := strconv.Atoi(submatches[1])
		if err != nil || idx < 0 || idx >= len(r.urlReports) {
			return match
		}
		reports = append(reports, r.urlReports[idx])
		return ""
	})

	return result, reports
}
