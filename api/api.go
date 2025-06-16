package api

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/mileusna/useragent"
	http "github.com/saucesteals/fhttp"
	"github.com/saucesteals/fhttp/cookiejar"

	"github.com/saucesteals/mimic"
)

var (
	ErrRateLimited = errors.New("rate limited")
	chromeVersion  = "137.0.0.0"
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
	userAgent useragent.UserAgent
}

func New(opts Options) (*API, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	transport, err := mimic.NewTransport(mimic.TransportOptions{
		Version:  chromeVersion,
		Brand:    mimic.BrandChrome,
		Platform: mimic.PlatformMac,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	})
	if err != nil {
		return nil, err
	}

	userAgent := useragent.Parse(transport.DefaultHeaders.Get("User-Agent"))
	if userAgent.String == "" {
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

func (a *API) GetUserAgent() useragent.UserAgent {
	return a.userAgent
}

func (a *API) GetCredentials() Credentials {
	return a.Credentials
}

func (a *API) Do(req *http.Request) (*http.Response, error) {
	res, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusTooManyRequests {
		res.Body.Close()
		return nil, ErrRateLimited
	}

	return res, nil
}
