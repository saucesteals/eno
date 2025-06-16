package extension

import (
	"context"
	"net/http"

	"github.com/saucesteals/eno/api"
)

func (a *Extension) CreateToken(ctx context.Context, tokenName string, card PaymentCard, merchant DataSource) (api.Token, error) {
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
		return api.Token{}, err
	}

	signature, err := a.sign(keys, "/defaultcard/tokenize")
	if err != nil {
		return api.Token{}, err
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

	var response api.Token
	req, err := a.newWibRequest(ctx, http.MethodPost, "token/defaultcard/tokenize", payload)
	if err != nil {
		return response, err
	}

	if err := a.do(req, &response); err != nil {
		return response, err
	}

	token, err := a.Decrypt(ctx, keys, response.Token)
	if err != nil {
		return api.Token{}, err
	}

	response.Token = token
	return response, nil
}
