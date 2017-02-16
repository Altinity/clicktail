/*
Package urlshaper creates a normalized shape (and other fields) from a URL.

Summary

Given a URL or a URL Path (eg http://example.com/foo/bar?q=baz or just
/foo/bar?q=baz) and an optional list of URL patterns, urlshaper will return an
object that has that URL broken up in to its various components and provide a
normalized shape of the URL.

Inputs

URL inputs to the urlshaper should be strings. They can be either fully
qualified URLs or just the path. (Anything that the net/url parser can parse
should be fine.)

Valid URL inputs:
    http://example.com/foo/bar
    https://example.com:8080/foo?q=bar
    /foo/bar/baz
    /foo?bar=baz

Patterns should describe only the path section of the URL. Variable portions of
the URL should be identified by a preceeding the section name with a colon
(":"). To match additional sections after the pattern, include a terminal
asterisk ("*")

Valid patterns:
    /about               matches /about and /about?q=bar
    /about/:lang         matches /about/en and /about/1234?q=bar
    /about/:lang/page    matches /about/en/page and /about/1234/page?q=bar
    /about/*             matches /about/foo/bar/baz and /about/a/b/c?q=bar

Output

If there is no error, the returned Result objected always has URI, Path, and
Shape filled in. The remaining fields will have zero values if the corresponding
sections of the URL are missing.

*/
package urlshaper
