package api

import (
	"net/url"

	http "github.com/saucesteals/fhttp"
)

var (
	wibURL = &url.URL{
		Scheme: "https",
		Host:   "wib.capitalone.com",
	}

	verifiedURL = &url.URL{
		Scheme: "https",
		Host:   "verified.capitalone.com",
	}

	accountsURL = &url.URL{
		Scheme: "https",
		Host:   "myaccounts.capitalone.com",
	}
)

func (a *API) GetCookies() []*http.Cookie {
	return a.jar.Cookies(verifiedURL)
}

func (a *API) GetCookie(name string) string {
	cookies := a.GetCookies()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func (a *API) SetCookies(cookies []*http.Cookie) {
	a.jar.SetCookies(
		wibURL,
		cookies,
	)
	a.jar.SetCookies(
		verifiedURL,
		cookies,
	)
	a.jar.SetCookies(
		accountsURL,
		cookies,
	)
}
