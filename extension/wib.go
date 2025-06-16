package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	http "github.com/saucesteals/fhttp"
)

func getEWAFingerprint() string {
	return fmt.Sprintf(`{"container":{"width":340,"height":300,"opacity":"1"},"frame":{"width":340,"height":520,"opacity":"1"},"date":"%d"}`, time.Now().UnixMilli())
}

func (a *Extension) newWibRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, "https://wib.capitalone.com/wib-edge-server/"+path, bodyReader)
	if err != nil {
		return nil, err
	}

	fingerprint, err := a.generateFingerprint()
	if err != nil {
		return nil, err
	}

	ua := a.api.GetUserAgent()
	req.Header.Add("accept", "application/json")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("cache-control", "no-cache, no-store, must-revalidate")
	req.Header.Add("dnt", "1")
	req.Header.Add("expires", "0")
	req.Header.Add("origin", GetChromeExtensionURL())
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("priority", "u=1, i")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-storage-access", "active")
	req.Header.Add("sec-gpc", "1")
	req.Header.Add("client-correlation-id", a.clientReferenceId)

	req.Header.Add("x-device-fingerprint", fingerprint)
	req.Header.Add("x-apptype", ua.Name)
	req.Header.Add("x-appversion", GetChromeExtensionVersion())
	req.Header.Add("x-browserversion", ua.Version)
	req.Header.Add("x-devicemodel", "Mac OS")
	req.Header.Add("x-osversion", ua.OSVersion)
	req.Header.Add("x-platform", "walletinbrowser")

	if path == "token/defaultcard/tokenize" {
		fingerprint, err := a.EncryptEWA(ctx, getEWAFingerprint())
		if err != nil {
			return nil, err
		}

		req.Header.Set("ewa-fingerprint", fingerprint)
	}

	if profileReferenceId := a.session.GetProfileReferenceId(); profileReferenceId != "" {
		req.Header.Set("profile_ref_id", profileReferenceId)
		req.Header.Set("profilereferenceid", profileReferenceId)
	}

	if accessToken := a.session.GetAccessToken(); accessToken != "" {
		req.Header.Set("access-token", accessToken)
	}

	if ipAddress := a.session.GetIpAddress(); ipAddress != "" {
		req.Header.Set("client-ip", ipAddress)
	}

	if bodyReader != nil {
		req.Header.Set("content-type", "application/json;charset=UTF-8")
	}

	return req, nil
}

func (a *Extension) do(req *http.Request, body any) error {
	res, err := a.api.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return fmt.Errorf("%s %s: status code: %d", req.Method, req.URL.String(), res.StatusCode)
	}

	if accessToken := res.Header.Get("access-token"); accessToken != "" {
		a.session.SetAccessToken(accessToken)
	}

	if ipAddress := res.Header.Get("client-ip"); ipAddress != "" {
		a.session.SetIpAddress(ipAddress)
	}

	if body != nil {
		return json.NewDecoder(res.Body).Decode(body)
	}

	return nil
}
