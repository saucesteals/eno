package api

import (
	"net/url"
	"time"

	http "github.com/saucesteals/fhttp"
)

var (
	cookieURL = &url.URL{
		Scheme: "https",
		Host:   ".capitalone.com",
	}
)

func (a *API) GetCookies() []*http.Cookie {
	cookies := a.jar.Cookies(cookieURL)
	for _, cookie := range cookies {
		cookie.Domain = cookieURL.Host
		cookie.Expires = time.Now().Add(time.Hour * 24 * 30)
	}

	return cookies
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
	for _, cookie := range cookies {
		cookie.Domain = cookieURL.Host
		cookie.Expires = time.Now().Add(time.Hour * 24 * 30)
	}

	a.jar.SetCookies(
		cookieURL,
		cookies,
	)
}
