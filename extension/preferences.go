package extension

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type PreferredCard struct {
	AssetLocationURL       string `json:"assetLocationUrl"`
	BrandColor             string `json:"brandColor"`
	CardReferenceID        string `json:"cardReferenceId"`
	CardDescription        string `json:"cardDescription"`
	CardFirstSix           string `json:"cardFirstSix"`
	CardLastFour           string `json:"cardLastFour"`
	CardNetworkType        string `json:"cardNetworkType"`
	CardLockState          string `json:"cardLockState"`
	CustomerName           string `json:"customerName"`
	CardImageCode          string `json:"cardImageCode"`
	EligibleCardsAvailable bool   `json:"eligibleCardsAvailable"`
	CardIneligible         bool   `json:"cardIneligible"`
	CardFraudLocked        bool   `json:"cardFraudLocked"`
	CardUserLocked         bool   `json:"cardUserLocked"`
}

func (a *Extension) GetPreferredCards(ctx context.Context) ([]PreferredCard, error) {
	type Response struct {
		PreferredCards []PreferredCard `json:"preferredCards"`
		Type           string          `json:"type"`
	}

	var response Response
	req, err := a.newWibRequest(ctx, http.MethodGet, fmt.Sprintf("wib/user/preferences?profileReferenceId=%s&shouldIncludeCardStatus=true", a.session.GetProfileReferenceId()), nil)
	if err != nil {
		return response.PreferredCards, err
	}

	if err := a.do(req, &response); err != nil {
		return response.PreferredCards, err
	}

	return response.PreferredCards, nil
}

func (a *Extension) ConfigureCard(ctx context.Context, card PaymentCard, cvv string) error {
	type Payload struct {
		PaymentCardReferenceID string `json:"paymentCardReferenceId"`
		UserEnteredCvv         string `json:"userEnteredCvv"`
		LastFour               string `json:"lastFour"`
		DeviceExtensionid      string `json:"deviceExtensionid"`
	}

	type Response struct {
		ResponseID   string `json:"responseID"`
		FirstSix     string `json:"firstSix"`
		LastFour     string `json:"lastFour"`
		WasEmailSent bool   `json:"wasEmailSent"`
	}
	if len(card.CardNumber) != 16 {
		return fmt.Errorf("invalid card number")
	}

	payload := Payload{
		PaymentCardReferenceID: card.CardReferenceID,
		UserEnteredCvv:         cvv,
		LastFour:               card.CardNumber[len(card.CardNumber)-4:],
		DeviceExtensionid:      a.device.ExtensionId,
	}

	var response Response
	req, err := a.newWibRequest(ctx, http.MethodPost, fmt.Sprintf("wib/user/preferences?profileReferenceId=%s&shouldIncludeCardStatus=true", a.session.GetProfileReferenceId()), payload)
	if err != nil {
		return err
	}

	if err := a.do(req, &response); err != nil {
		return err
	}

	if response.WasEmailSent {
		return fmt.Errorf("email sent on card configuration")
	}

	return nil
}

func (a *Extension) CompleteCardRegister(ctx context.Context, card PaymentCard) error {
	type Payload struct {
		AccountRefID  string `json:"accountRefId"`
		CardLastFour  string `json:"cardLastFour"`
		CustFirstName string `json:"custFirstName"`
	}

	type Response struct {
		IsSuccess bool `json:"isSuccess"`
	}

	names := strings.Split(card.CustomerName, " ")
	if len(card.CardNumber) != 16 {
		return fmt.Errorf("invalid card number")
	}

	payload := Payload{
		AccountRefID:  card.AccountReferenceID,
		CardLastFour:  card.CardNumber[len(card.CardNumber)-4:],
		CustFirstName: names[0],
	}

	var response Response
	req, err := a.newWibRequest(ctx, http.MethodPost, "wib/email/complete-register", payload)
	if err != nil {
		return err
	}

	if err := a.do(req, &response); err != nil {
		return err
	}

	if !response.IsSuccess {
		return fmt.Errorf("failed to complete register")
	}

	return nil
}
