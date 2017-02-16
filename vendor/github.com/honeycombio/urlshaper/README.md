# urlshaper

URL Shaper is a go library that takes a URL path and query and breaks it up into its components and creates a shape.

URL Shaper will generate a URL Shape by taking a URL and making a number of changes to it to make it easier to do analysis on URL patterns:
- replaces query parameter values with a question mark
- query parameters in the URL shape are alphabetized
- based on patterns, variables in the URL path are replaced by the variable name

In addition to the URL shape, the Result object you get back from URL Shaper contains
- the path, with any query parameters removed
- the query, with the path removed
- a url.Values object containing all the query parameters
- a url.Values object containing all the path parameters

Parameters in the path portion of the URL are identified by matching the URL against a list of provided patterns. Patterns are matched in the order provided; the first match wins. Patterns should represent the entire path portion of the URL - include a "*" at the end to match arbitrary additional segments.

## Examples

### Query parameters:

input:

- path: `/about/en/books?isbn=123456&author=Alice`

output:

- uri: `/about/en/books?isbn=123456&author=Alice`
- path: `/about/en/books`
- query: `isbn=123456&author=Alice`
- query_fields: `{"isbn":["123456"],"author":["Alice"]}`
- path_fields: `{}`
- shape: `/about/en/books?author=?&isbn=?`

### REST:

input:

- path: `/about/en/books/123456`
- pattern: `/about/:lang/books/:isbn`

output:

- uri: `/about/en/books/123456`
- path: `/about/en/books/123456`
- query: `""`
- query_fields: `{}`
- path_fields: `{"lang":["en"],"isbn":["123456"]}`
- shape: `/about/:lang/books/:isbn`

### REST & Query parameters:

input:

- path: `/about/en/books?isbn=123456&author=Alice&isbn=987654`
- pattern: `/about/:lang/books`

output:

- uri: `/about/en/books?isbn=123456&author=Alice`
- path: `/about/en/books`
- query: `isbn=123456&author=Alice`
- query_fields: `{"isbn":["123456", "987654"],"author":["Alice"]}`
- path_fields: `{"lang":["en"]}`
- shape: `/about/:lang/books?author=?&isbn=?`

### Unmatched:

input:

- path `/other/path`
- patterns: `/about/:lang/books`, `/docs/:section`

output:

- uri: `/other/path`
- path: `/other/path`
- query: `""`
- query_fields: `{}`
- path_fields: `{}`
- shape: `/other/path`

### Wildcard:

input:

-path `/docs/quickstart/linux`
-pattern `/docs/:section/*`

output:

- uri: `/docs/quickstart/linux`
- path: `/docs/quickstart/linux`
- query: `""`
- query_fields: `{}`
- path_fields: `{"quickstart":["linux"]}`
- shape: `/docs/:section/*`
