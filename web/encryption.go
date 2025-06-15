package web

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"

	http "github.com/saucesteals/fhttp"

	"github.com/go-jose/go-jose/v3"
)

type ProductId string

const (
	ProductIdProd ProductId = "gwlite-ease-prod"
	ProductIdCDE  ProductId = "gwlite-ease-cde"
)

func (a *Web) GetGWLiteKey(ctx context.Context, productId ProductId) (*jose.JSONWebKey, error) {
	req, err := a.newWebRequest(ctx, http.MethodGet, fmt.Sprintf("oidc/key-management/certificates/keys?productId=%s&use=enc&getAllKeys=false&kty=RSA", productId), nil, nil)
	if err != nil {
		return nil, err
	}

	res, err := a.api.Do(req)
	if err != nil {
		return nil, err
	}

	var response jose.JSONWebKeySet
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Keys) == 0 {
		return nil, errors.New("no gwlite keys found")
	}

	return &response.Keys[0], nil
}

func (a *Web) GenerateJWK(ctx context.Context) (*jose.JSONWebKey, *rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	return &jose.JSONWebKey{Key: &key.PublicKey}, key, nil
}
