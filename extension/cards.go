package extension

import (
	"context"
	"net/http"
)

type PaymentCard struct {
	AssetLocationURL            string `json:"assetLocationUrl"`
	BrandColor                  string `json:"brandColor"`
	AccountReferenceID          string `json:"accountReferenceId"`
	CardNumber                  string `json:"cardNumber"`
	PaymentCardType             string `json:"paymentCardType"`
	CardReferenceID             string `json:"cardReferenceId"`
	CustomerName                string `json:"customerName"`
	CardHolderRole              string `json:"cardHolderRole"`
	CustomerReferenceID         string `json:"customerReferenceId"`
	CardStatus                  string `json:"cardStatus"`
	CardNetworkType             string `json:"cardNetworkType"`
	ProductDescription          string `json:"productDescription"`
	CardTypeID                  string `json:"cardTypeId"`
	CardImageCode               string `json:"cardImageCode"`
	IsCardLocked                bool   `json:"isCardLocked"`
	IsFraudLocked               bool   `json:"isFraudLocked"`
	IsCvvLocked                 bool   `json:"isCvvLocked"`
	IsCvvValidationRequired     bool   `json:"isCvvValidationRequired"`
	IsProvisioned               bool   `json:"isProvisioned"`
	DefaultPaymentCardIndicator bool   `json:"defaultPaymentCardIndicator"`
}

func (a *Extension) GetPaymentCards(ctx context.Context) ([]PaymentCard, error) {
	type Response struct {
		PaymentCards          []PaymentCard `json:"paymentCards"`
		HasEligibleCards      bool          `json:"hasEligibleCards"`
		IsAllCardsFraudLocked bool          `json:"isAllCardsFraudLocked"`
		IsAllCardsCVVlocked   bool          `json:"isAllCardsCVVlocked"`
		IsAllCardsUserLocked  bool          `json:"isAllCardsUserLocked"`
	}

	var response Response
	req, err := a.newWibRequest(ctx, http.MethodGet, "wib/payment-cards", nil)
	if err != nil {
		return nil, err
	}

	if err := a.do(req, &response); err != nil {
		return nil, err
	}

	return response.PaymentCards, nil
}
