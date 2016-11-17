package parsers

import "regexp"

// ExtRegexp is a Regexp with one additional method to make it easier to work
// with named groups
type ExtRegexp struct {
	*regexp.Regexp
}

// FindStringSubmatchMap behaves the same as FindStringSubmatch except instead
// of a list of matches with the names separate, it returns the full match and a
// map of named submatches
func (r *ExtRegexp) FindStringSubmatchMap(s string) (string, map[string]string) {
	match := r.FindStringSubmatch(s)
	if match == nil {
		return "", nil
	}

	captures := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i == 0 {
			continue
		}
		if name != "" {
			// ignore unnamed matches
			captures[name] = match[i]
		}
	}
	return match[0], captures
}
