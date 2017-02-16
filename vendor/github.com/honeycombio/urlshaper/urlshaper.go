package urlshaper

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// Result contains the parsed portions of the input URL
type Result struct {
	// URI is the original string that is parsed
	URI string
	// Path is the unmodified path portion of the URL
	Path string
	// Query is the unmodified query portion of the URL
	Query string
	// QueryFields is a map of the query parameters to values
	QueryFields url.Values
	// PathFields is a map of the keys in the provided pattern to the values
	// extracted from the Path
	PathFields url.Values
	// Shape is the normalized URL, with all variable portions replaced by '?'
	// and query parameters sorted alphabetically
	Shape string
	// PathShape is the path portion of the normalized URL
	PathShape string
	// QueryShape is the query portion of the normalized URL
	QueryShape string
}

// Pattern is an object that represents a URL path pattern you wish to use when
// creating the shape. After adding a pattern to the Pattern object, you should
// call Compile on the pattern so it is suitable to use for matching.
//
// If you don't call Compile, the pattern will be automatically compiled, but
// you won't see any errors that come up during compilation. If the compile
// fails, the pattern will be silently ignored.
//
// Patterns should not include any query parameters - only the path portion of
// the URL is tested against the pattern.
type Pattern struct {
	Pat string
	// rawRE is the raw regex translated from the pattern. saved for testing
	rawRE string
	//cpat is the compiled pattern for use in matching
	re *regexp.Regexp
	// if we failed to compile (rather than haven't yet been compiled), set this
	// flag so we can skip trying to compile again
	failed bool
}

// Compile turns the Pattern into a compiled regex that will be used to match
// URL patterns
func (p *Pattern) Compile() error {
	if p.re != nil {
		// we've already been compiled!
		return nil
	}
	rawRE := p.Pat
	varRE, err := regexp.Compile("/(:[^/]+)")
	if err != nil {
		p.failed = true
		return err
	}
	vars := varRE.FindAllStringSubmatch(p.Pat, -1)
	for _, varMatch := range vars {
		varName := strings.Trim(varMatch[0], "/:")
		regName := fmt.Sprintf("/(?P<%s>[^/]+)", varName)
		rawRE = strings.Replace(rawRE, varMatch[0], regName, 1)
	}
	// handle * at the end of the pattern
	if strings.HasSuffix(rawRE, "*") {
		rawRE = strings.TrimSuffix(rawRE, "*")
		rawRE += ".*"
	}
	// anchor the regex to both the beginning and the end
	rawRE = "^" + rawRE + "$"
	p.rawRE = rawRE

	re, err := regexp.Compile(rawRE)
	if err != nil {
		p.failed = true
		return err
	}
	p.re = re
	return nil
}

// Parser contains a list of Patterns to use for parsing URLs, then exposes the
// functionality to generate a Result.
type Parser struct {
	Patterns []*Pattern
}

// Parse takes a URL string and a list of patterns, attempts to parse, and
// and hands back a Result and error
func (p *Parser) Parse(rawURL string) (*Result, error) {
	// split out the path and query
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	r := &Result{
		URI:         rawURL,
		Path:        u.Path,
		Query:       u.RawQuery,
		QueryFields: u.Query(),
		PathFields:  url.Values{},
	}

	// create the shaped query string
	var queryShape string
	// sort the query parameters
	qKeys := make([]string, 0, len(r.QueryFields))
	for k := range r.QueryFields {
		qKeys = append(qKeys, k)
	}
	sort.Strings(qKeys)

	shapeStrings := make([]string, 0, len(qKeys))
	for _, k := range qKeys {
		for i := 0; i < len(r.QueryFields[k]); i++ {
			shapeStrings = append(shapeStrings, fmt.Sprintf("%s=?", k))
		}
	}
	queryShape = strings.Join(shapeStrings, "&")

	// try and find a shape for the path or use the Path as the shape
	var pathShape string
	pathShape, r.PathFields = p.getPathShape(u.Path)

	// Ok, construct the full shape of the URL
	fullShape := pathShape
	if queryShape != "" {
		fullShape += fmt.Sprintf("?%s", queryShape)
	}
	r.PathShape = pathShape
	r.QueryShape = queryShape
	r.Shape = fullShape

	return r, nil
}

// getPathShape takes a path and does all the pattern matching and extraction
// of path variables
func (p *Parser) getPathShape(path string) (string, url.Values) {
	pathShape := path
	pathFields := url.Values{}
	for _, pat := range p.Patterns {
		if pat.failed {
			// skip failed patterns
			continue
		}
		// try and compile not-yet-compiled patterns
		if pat.re == nil {
			if err := pat.Compile(); err != nil {
				continue
			}
		}
		matches := pat.re.FindStringSubmatch(path)
		if matches == nil {
			// no match, try the next pattern
			continue
		}
		pathShape = pat.Pat
		if len(matches) == 1 {
			// no variables in the pattern. All done
			break
		}
		names := pat.re.SubexpNames()
		for i, val := range matches {
			if i == 0 {
				// skip the full string match
				continue
			}
			pathFields.Add(names[i], val)
		}
		// since we matched, don't check any more patterns in the list
		break
	}
	return pathShape, pathFields
}
