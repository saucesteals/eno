package api

import (
	"fmt"
	"log/slog"

	http "github.com/saucesteals/fhttp"
	"github.com/saucesteals/fhttp/cookiejar"

	"github.com/saucesteals/mimic"
)

type Credentials struct {
	Username string
	Password string
}

type Options struct {
	Logger *slog.Logger

	Credentials         Credentials
	BrowserUserDataPath string
	BrowserBinary       string
}

type API struct {
	Options

	client    *http.Client
	jar       *cookiejar.Jar
	userAgent string
}

func New(opts Options) (*API, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	transport, err := mimic.NewTransport(mimic.TransportOptions{
		Version:  "137.0.0.0",
		Brand:    mimic.BrandChrome,
		Platform: mimic.PlatformMac,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	})
	if err != nil {
		return nil, err
	}

	userAgent := transport.DefaultHeaders.Get("User-Agent")
	if userAgent == "" {
		return nil, fmt.Errorf("no user agent found")
	}

	a := &API{
		Options: opts,

		client: &http.Client{
			Transport: transport,
			Jar:       jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		jar:       jar,
		userAgent: userAgent,
	}

	return a, nil
}

func (a *API) GetUserAgent() string {
	return a.userAgent
}

func (a *API) GetCredentials() Credentials {
	return a.Credentials
}

func (a *API) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", a.userAgent)
	return a.client.Do(req)
}
