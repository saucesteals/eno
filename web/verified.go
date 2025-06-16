package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/google/uuid"
	http "github.com/saucesteals/fhttp"
)

func (a *Web) newVerifiedRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, "https://verified.capitalone.com/"+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept-Language", "en-us")
	req.Header.Add("ui-version", "12")
	req.Header.Add("api-key", "RTM")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	// req.Header.Add("DPoP", "...")
	req.Header.Add("accept", "application/json;v=3")
	req.Header.Add("content-type", "application/json;v=3")
	req.Header.Add("identity_channel_type", "desktop")
	req.Header.Add("Client-Correlation-Id", a.getVerifiedClientCorrelationId())
	req.Header.Add("Origin", "https://verified.capitalone.com")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Referer", "https://verified.capitalone.com/")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br, zstd")

	return req, nil
}

func generateVerifiedClientCorrelationId() string {
	return "UV12-SIC-" + uuid.NewString()
}

// thread unsafe
func (a *Web) getVerifiedClientCorrelationId() string {
	if a.verifiedCCId != "" {
		return a.verifiedCCId
	}

	a.verifiedCCId = generateVerifiedClientCorrelationId()
	return a.verifiedCCId
}
