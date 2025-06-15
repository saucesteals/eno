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

	req.Header = http.Header{
		"accept":                   {"application/json"},
		"accept-language":          {"en-US,en;q=0.9"},
		"cache-control":            {"no-cache, no-store, must-revalidate"},
		"dnt":                      {"1"},
		"expires":                  {"0"},
		"origin":                   {GetChromeExtensionURL()},
		"pragma":                   {"no-cache"},
		"priority":                 {"u=1, i"},
		"sec-fetch-dest":           {"empty"},
		"sec-fetch-mode":           {"cors"},
		"sec-fetch-site":           {"none"},
		"sec-fetch-storage-access": {"active"},
		"sec-gpc":                  {"1"},
		"client-correlation-id":    {a.clientReferenceId},
		"User-Agent":               {a.device.UserAgent.String},
		"x-device-fingerprint":     {a.fingerprint},
		"x-apptype":                {a.device.UserAgent.Name},
		"x-appversion":             {GetChromeExtensionVersion()},
		"x-browserversion":         {a.device.UserAgent.Version},
		"x-devicemodel":            {"Mac OS"},
		"x-osversion":              {a.device.UserAgent.OSVersion},
		"x-platform":               {"walletinbrowser"},
	}

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
