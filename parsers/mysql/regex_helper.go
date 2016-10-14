package mysql

import "regexp"

type myRegexp struct {
	*regexp.Regexp
}

var myExp = myRegexp{regexp.MustCompile(`(?P<first>\d+)\.(?P<second>\d+)`)}

func (r *myRegexp) FindStringSubmatchMap(s string) map[string]string {
	match := r.FindStringSubmatch(s)
	if match == nil {
		return nil
	}

	captures := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i == 0 {
			continue
		}
		captures[name] = match[i]
	}
	return captures
}
