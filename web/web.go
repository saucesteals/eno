package web

import (
	"bytes"
	"context"

	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"

	http "github.com/saucesteals/fhttp"

	"github.com/go-jose/go-jose/v3"
	"github.com/saucesteals/eno/api"
)

type Web struct {
	api *api.API
}

func New(api *api.API) *Web {
	return &Web{
		api: api,
	}
}

func newSynchToken() string {
	now := []byte(strconv.FormatInt(time.Now().UnixMilli(), 10))
	for i := len(now) - 1; i >= 0; i-- {
		j := mrand.Intn(i + 1)
		now[i], now[j] = now[j], now[i]
	}

	return string(now)
}

func (a *Web) newWebRequest(ctx context.Context, method, path string, body any, key *jose.JSONWebKey) (*http.Request, error) {
	isProtected := strings.HasPrefix(path, "web-api/tiger/protected")
	isOidc := strings.HasPrefix(path, "oidc/")

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		if isProtected {
			serverKey, err := a.GetGWLiteKey(ctx, ProductIdCDE)
			if err != nil {
				return nil, err
			}

			encrypter, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{
				Algorithm: jose.RSA_OAEP_256,
				Key:       serverKey,
			}, nil)
			if err != nil {
				return nil, err
			}

			encrypted, err := encrypter.Encrypt(payload)
			if err != nil {
				return nil, err
			}

			serialized, err := encrypted.CompactSerialize()
			if err != nil {
				return nil, err
			}

			bodyReader = bytes.NewReader([]byte(serialized))
		} else {
			bodyReader = bytes.NewReader(payload)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, "https://myaccounts.capitalone.com/"+path, bodyReader)
	if err != nil {
		return nil, err
	}

	synchToken := newSynchToken()
	req.Header = http.Header{
		"accept-language":          {"en-US,en;q=0.9"},
		"cache-control":            {"no-cache, no-store, must-revalidate"},
		"dnt":                      {"1"},
		"expires":                  {"0"},
		"origin":                   {"https://myaccounts.capitalone.com"},
		"pragma":                   {"no-cache"},
		"priority":                 {"u=1, i"},
		"sec-fetch-dest":           {"empty"},
		"sec-fetch-mode":           {"cors"},
		"sec-fetch-site":           {"none"},
		"sec-fetch-storage-access": {"active"},
		"sec-gpc":                  {"1"},
		"evt_synch_token":          {synchToken},
	}

	if isProtected {
		if key == nil {
			return nil, errors.New("key is required")
		}
		req.Header.Set("accept", "application/jwt;v=1")

		serialized, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}

		encoded := base64.StdEncoding.EncodeToString(serialized)
		req.Header.Set("x-gw-client-public-key", string(encoded))
	} else if isOidc {
		req.Header.Set("accept", "application/json, text/plain, */*")
	} else {
		req.Header.Set("accept", "application/json;v=2")
	}

	if bodyReader != nil {
		if isProtected {
			req.Header.Set("content-type", "application/jwt")
		} else {
			req.Header.Set("content-type", "application/json")
		}
	}

	return req, nil
}

func (a *Web) do(req *http.Request, body any, key *rsa.PrivateKey) error {
	res, err := a.api.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return fmt.Errorf("%s %s: status code: %d", req.Method, req.URL.String(), res.StatusCode)
	}

	if body != nil {
		if key == nil {
			return json.NewDecoder(res.Body).Decode(body)
		}

		response, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, response, []byte(""))
		if err != nil {
			return err
		}

		err = json.Unmarshal(decrypted, body)
		if err != nil {
			return err
		}
	}

	return nil
}
