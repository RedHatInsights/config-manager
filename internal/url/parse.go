package url

import "net/url"

// MustParse attempts to parse rawurl using the net/url package, and panics in
// the event of an error.
func MustParse(rawurl string) *url.URL {
	URL, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return URL
}
