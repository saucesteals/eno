package web

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/saucesteals/eno/api"
	"github.com/saucesteals/eno/extension"
)

func (a *Web) CreateToken(ctx context.Context, tokenName string, card extension.PaymentCard) (api.Token, error) {
	type Payload struct {
		CardReferenceID string `json:"cardReferenceId"`
		IsOneTimeUse    bool   `json:"isOneTimeUse"`
		MdxID           string `json:"mdxId"`
		MdxURLID        string `json:"mdxUrlId"`
		TokenCardName   string `json:"tokenCardName"`
		TokenDuration   string `json:"tokenDuration"`
	}

	payload := Payload{
		CardReferenceID: card.CardReferenceID,
		MdxID:           "",
		MdxURLID:        "",
		TokenCardName:   tokenName,
		IsOneTimeUse:    false,
		TokenDuration:   "",
	}

	clientKey, clientPrivateKey, err := a.GenerateJWK(ctx)
	if err != nil {
		return api.Token{}, err
	}

	req, err := a.newWebRequest(ctx, http.MethodPost, "web-api/tiger/protected/222543/commerce-virtual-numbers", payload, clientKey)
	if err != nil {
		return api.Token{}, err
	}

	var token api.Token
	if err := a.do(req, &token, clientPrivateKey); err != nil {
		return api.Token{}, err
	}

	return token, nil
}

type UpdateTokenPayload struct {
	AllowAuthorizations bool   `json:"allowAuthorizations"`
	CardLastFour        string `json:"cardLastFour"`
	CardName            string `json:"cardName"`
	CardReferenceID     string `json:"cardReferenceId"`
	IsDeleted           bool   `json:"isDeleted"`
	MdxID               string `json:"mdxId"`
	MdxURLID            string `json:"mdxUrlId"`
	TokenDuration       any    `json:"tokenDuration"`
	TokenLastFour       string `json:"tokenLastFour"`
	TokenName           string `json:"tokenName"`
	TokenReferenceID    string `json:"tokenReferenceId"`
}

func (a *Web) UpdateToken(ctx context.Context, update UpdateTokenPayload) error {
	req, err := a.newWebRequest(ctx, http.MethodPut, "web-api/private/25419/commerce-virtual-numbers", update, nil)
	if err != nil {
		return err
	}

	if err := a.do(req, nil, nil); err != nil {
		return err
	}

	return nil
}

type MdxInfo struct {
	MdxURLID      string `json:"mdxUrlId"`
	MerchantURL   string `json:"merchantUrl"`
	Name          string `json:"name"`
	MdxID         string `json:"mdxId"`
	Color         any    `json:"color"`
	ColorContrast any    `json:"colorContrast"`
	ImageURL      any    `json:"imageUrl"`
}

type ListedToken struct {
	TokenUpdatedTimestamp        string  `json:"tokenUpdatedTimestamp"`
	DerivedStatus                string  `json:"derivedStatus"`
	TokenReferenceID             string  `json:"tokenReferenceId"`
	TokenName                    string  `json:"tokenName"`
	FormattedTokenExpirationDate string  `json:"formattedTokenExpirationDate"`
	CardReferenceID              string  `json:"cardReferenceId"`
	TokenCreatedTimestamp        string  `json:"tokenCreatedTimestamp"`
	TokenType                    string  `json:"tokenType"`
	HasTokenExpired              bool    `json:"hasTokenExpired"`
	TokenLastFour                string  `json:"tokenLastFour"`
	DurationHasPassed            bool    `json:"durationHasPassed"`
	TokenStatus                  string  `json:"tokenStatus"`
	MdxInfo                      MdxInfo `json:"mdxInfo"`
}

type ListTokensResponse struct {
	Entries         []ListedToken `json:"entries"`
	Limit           int           `json:"limit"`
	Offset          int           `json:"offset"`
	Count           int           `json:"count"`
	CachedCount     int           `json:"cachedCount"`
	UnfilteredCount int           `json:"unfilteredCount"`
}

func (a *Web) ListTokens(ctx context.Context, card extension.PaymentCard, nameFilter string, offset int, limit int) (ListTokensResponse, error) {
	type FilterCriteria struct {
		Field    string `json:"field"`
		Operator string `json:"operator"`
		Value    string `json:"value"`
	}

	type Payload struct {
		FilterCriteria  []FilterCriteria `json:"filterCriteria"`
		ReferenceId     string           `json:"referenceId"`
		ReferenceIdType string           `json:"referenceIdType"`
		SortCriteria    []any            `json:"sortCriteria"`
		TokenStatus     []any            `json:"tokenStatus"`
	}

	query := url.Values{}
	query.Add("offset", strconv.Itoa(offset))
	query.Add("limit", strconv.Itoa(limit))
	query.Add("excludeUnbound", "true")

	payload := Payload{
		ReferenceId:     card.CardReferenceID,
		ReferenceIdType: "ACCOUNT",

		FilterCriteria: []FilterCriteria{},
		SortCriteria:   []any{},
		TokenStatus:    []any{},
	}

	if nameFilter != "" {
		payload.FilterCriteria = append(payload.FilterCriteria, FilterCriteria{
			Field:    "TOKEN_NAME",
			Operator: "LIKE",
			Value:    nameFilter,
		})
	}

	req, err := a.newWebRequest(ctx, http.MethodPost, "web-api/private/25419/commerce-virtual-numbers?"+query.Encode(), payload, nil)
	if err != nil {
		return ListTokensResponse{}, err
	}

	var response ListTokensResponse
	if err := a.do(req, &response, nil); err != nil {
		return ListTokensResponse{}, err
	}

	return response, nil
}
