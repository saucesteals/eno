package extension

import (
	"context"
	"net/http"
)

type TokenMerchantBinding struct {
	BindingType  string `json:"bindingType"`
	MdxID        string `json:"mdxId"`
	MerchantName any    `json:"merchantName"`
	URLID        string `json:"urlId"`
}

type TokenRules struct {
	AllowAuthorizations bool                 `json:"allowAuthorizations"`
	MerchantBinding     TokenMerchantBinding `json:"merchantBinding"`
}

type Token struct {
	Token            string     `json:"token"`
	Cvv              string     `json:"cvv"`
	ExpirationDate   string     `json:"expirationDate"`
	LastFour         string     `json:"lastFour"`
	TokenReferenceID string     `json:"tokenReferenceId"`
	CreatedTimestamp string     `json:"createdTimestamp"`
	TokenName        string     `json:"tokenName"`
	TokenStatus      string     `json:"tokenStatus"`
	TokenType        string     `json:"tokenType"`
	TokenRules       TokenRules `json:"tokenRules"`
	CardReferenceID  string     `json:"cardReferenceId"`
}

func (a *Extension) CreateToken(ctx context.Context, card PaymentCard, merchant DataSource, tokenName string) (Token, error) {
	type Payload struct {
		CardReferenceID    string    `json:"cardReferenceId"`
		DeviceExtensionid  string    `json:"deviceExtensionid"`
		MdxID              string    `json:"mdxId"`
		MdxURLID           string    `json:"mdxUrlId"`
		ProfileReferenceID string    `json:"profileReferenceId"`
		TokenCardName      string    `json:"tokenCardName"`
		Signature          Signature `json:"signature"`
	}

	keys, err := a.GenerateKeys(ctx)
	if err != nil {
		return Token{}, err
	}

	signature, err := a.sign(keys, "/defaultcard/tokenize")
	if err != nil {
		return Token{}, err
	}

	payload := Payload{
		CardReferenceID:    card.CardReferenceID,
		DeviceExtensionid:  a.device.ExtensionId,
		MdxID:              merchant.MDXId,
		MdxURLID:           merchant.MDXUrlId,
		ProfileReferenceID: a.session.GetProfileReferenceId(),
		TokenCardName:      tokenName,
		Signature:          signature,
	}

	var response Token
	req, err := a.newWibRequest(ctx, http.MethodPost, "token/defaultcard/tokenize", payload)
	if err != nil {
		return response, err
	}

	if err := a.do(req, &response); err != nil {
		return response, err
	}

	token, err := a.Decrypt(ctx, keys, response.Token)
	if err != nil {
		return Token{}, err
	}

	response.Token = token
	return response, nil
}
