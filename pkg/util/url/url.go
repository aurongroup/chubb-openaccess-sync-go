package url

import "net/url"

func IsValid(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	return u.Host != ""
}
