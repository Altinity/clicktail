package mysql

import "regexp"

type myRegexp struct {
	*regexp.Regexp
}

var myExp = myRegexp{regexp.MustCompile(`(?P<first>\d+)\.(?P<second>\d+)`)}

func (r *myRegexp) FindStringSubmatchMap(s string) map[string]string {
	captures := make(map[string]string)

	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}

	for i, name := range r.SubexpNames() {
		//
		if i == 0 {
			continue
		}
		captures[name] = match[i]

	}
	return captures
}
