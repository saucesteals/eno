package api

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"
	"unsafe"

	http "github.com/saucesteals/fhttp"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

var (
	loginURL = "https://api.capitalone.com/oauth2/authorize?client_id=22a835dd1466b71dab66c9e5ee3cbcf1&response_type=code&scope=openid&redirect_uri=https://verified.capitalone.com/sign-in/pathfinder"
)

func (a *API) initBrowser() (*rod.Browser, error) {
	ctx := context.Background()
	l := launcher.New().
		Context(ctx).
		Headless(false).
		UserDataDir(a.BrowserUserDataPath).
		Bin(a.BrowserBinary)

	controlUrl, err := l.Launch()
	if err != nil {
		return nil, err
	}

	browser := rod.New().Context(ctx).ControlURL(controlUrl)

	if err := browser.Connect(); err != nil {
		return nil, err
	}

	cookies := a.GetCookies()
	var browserCookies []*proto.NetworkCookieParam
	for _, cookie := range cookies {
		browserCookies = append(browserCookies, &proto.NetworkCookieParam{
			Name:   cookie.Name,
			Value:  cookie.Value,
			Domain: ".capitalone.com",
		})
	}

	return browser, browser.SetCookies(browserCookies)
}

func (a *API) initRouter(browser *rod.Browser, signIn chan error) (*rod.HijackRouter, error) {
	router := browser.HijackRequests()

	err := router.Add("*", "", func(h *rod.Hijack) {
		u := h.Request.URL()
		if u.Host != "verified.capitalone.com" && u.Host != "wib.capitalone.com" {
			h.ContinueRequest(&proto.FetchContinueRequest{})
			return
		}

		stdRequest := h.Request.Req()

		req, err := http.NewRequest(stdRequest.Method, stdRequest.URL.String(), stdRequest.Body)
		if err != nil {
			h.Response.Fail(proto.NetworkErrorReasonFailed)
			a.Logger.Error("new hijacked request", "url", stdRequest.URL.String(), "error", err)
			return
		}
		req.Header = http.Header(stdRequest.Header)

		res, err := a.client.Do(req)
		if err != nil {
			h.Response.Fail(proto.NetworkErrorReasonFailed)
			a.Logger.Error("do hijacked request", "url", stdRequest.URL.String(), "error", err)
			return
		}
		defer res.Body.Close()

		// Set the response code on the private :( payload field
		responseValue := reflect.ValueOf(h.Response).Elem()
		payloadField := responseValue.FieldByName("payload")
		if payloadField.IsValid() {
			payloadField = reflect.NewAt(payloadField.Type(), unsafe.Pointer(payloadField.UnsafeAddr())).Elem()
			if payloadField.CanInterface() {
				payload := payloadField.Interface().(*proto.FetchFulfillRequest)
				payload.ResponseCode = res.StatusCode
			} else {
				a.Logger.Error("Cannot interface with internal payload field")
			}
		} else {
			a.Logger.Error("Internal payload field is invalid")
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			h.Response.Fail(proto.NetworkErrorReasonFailed)
			a.Logger.Error("read response body", "url", stdRequest.URL.String(), "error", err)
			return
		}

		h.Response.SetBody(body)
		for k, vs := range res.Header {
			for _, v := range vs {
				h.Response.SetHeader(k, v)
			}
		}

		if u.Path != "/signincontroller-web/signincontroller/signin" {
			return
		}

		if res.StatusCode != http.StatusOK {
			signIn <- fmt.Errorf("status code: %d", res.StatusCode)
			return
		}

		var w1bCookies []*http.Cookie
		for _, cookie := range res.Cookies() {
			w1bCookies = append(w1bCookies, &http.Cookie{
				Name:       cookie.Name,
				Value:      cookie.Value,
				Path:       cookie.Path,
				Expires:    cookie.Expires,
				RawExpires: cookie.RawExpires,
				MaxAge:     cookie.MaxAge,
				Secure:     cookie.Secure,
				HttpOnly:   cookie.HttpOnly,
				Raw:        cookie.Raw,
				Unparsed:   cookie.Unparsed,
				SameSite:   http.SameSiteNoneMode,
				Domain:     ".capitalone.com",
			})
		}

		a.SetCookies(w1bCookies)
		signIn <- nil
	})

	go router.Run()

	return router, err
}

func (a *API) Login(ctx context.Context) error {
	browser, err := a.initBrowser()
	if browser != nil {
		defer browser.Close()
	}
	if err != nil {
		return err
	}

	signIn := make(chan error)
	router, err := a.initRouter(browser, signIn)
	if router != nil {
		defer router.Stop()
	}
	if err != nil {
		return err
	}

	page, err := browser.Page(proto.TargetCreateTarget{
		URL: loginURL,
	})
	if err != nil {
		return err
	}
	defer page.Close()

	typeElement := func(selector string, text string) error {
		element, err := page.Element(selector)
		if err != nil {
			return err
		}

		for _, char := range text {
			err := element.Input(string(char))
			if err != nil {
				return err
			}

			time.Sleep(100 * time.Millisecond)
		}

		return nil
	}

	clickElement := func(selector string) error {
		_, err := page.Eval(`(selector) => {
			const element = document.querySelector(selector);
			if (element) {
				element.click();
			}
		}`, selector)
		if err != nil {
			return err
		}

		return nil
	}

	err = typeElement(`input[data-testtarget="username-usernameInputField"]`, a.Credentials.Username)
	if err != nil {
		return err
	}

	err = typeElement(`#pwInputField`, a.Credentials.Password)
	if err != nil {
		return err
	}

	err = clickElement(`input[type="checkbox"]`)
	if err != nil {
		return err
	}

	err = clickElement(`button[data-testtarget="sign-in-submit-button"]`)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-signIn:
		return err
	}
}
