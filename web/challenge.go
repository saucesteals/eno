package web

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-jose/go-jose/v3"
	"github.com/saucesteals/eno/extension"
	http "github.com/saucesteals/fhttp"

	"github.com/google/uuid"
)

var (
	verificationBusinessEvent = "CARD.SERVICING.WEB.EASE.VIRTUAL_CARD_VCNCREATE"
)

func (a *Web) ChallengeValidation(ctx context.Context, policyProcessID string, otp ChallengeVerificationOtp, otpValue string) error {
	type Payload struct {
		BusinessEvent                 string `json:"businessEvent"`
		ChallengeMethod               string `json:"challengeMethod"`
		PolicyProcessID               string `json:"policyProcessId"`
		PassphraseAuthenticationToken string `json:"passphraseAuthenticationToken"`
		EncryptedPassphrase           string `json:"encryptedPassphrase"`
	}

	type Response struct {
		Authenticator string `json:"authenticator"`
		Otp           struct {
			AcceptanceStatus  string `json:"acceptanceStatus"`
			ProfileStatus     string `json:"profileStatus"`
			RemainingAttempts int    `json:"remainingAttempts"`
		} `json:"otp"`
		RedirectUrl string `json:"redirectUrl"`
	}

	decoded, err := base64.StdEncoding.DecodeString(otp.EncryptionKey)
	if err != nil {
		return err
	}

	var keys jose.JSONWebKeySet
	if err := json.Unmarshal(decoded, &keys); err != nil {
		return err
	}

	if len(keys.Keys) == 0 {
		return errors.New("no keys found")
	}

	encrypter, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{
		Algorithm: jose.ECDH_ES_A128KW,
		Key:       &keys.Keys[0],
	}, nil)
	if err != nil {
		return err
	}

	jwe, err := encrypter.Encrypt([]byte(otpValue))
	if err != nil {
		return err
	}

	encrypted, err := jwe.CompactSerialize()
	if err != nil {
		return err
	}

	payload := Payload{
		BusinessEvent:                 verificationBusinessEvent,
		ChallengeMethod:               "OTP",
		PolicyProcessID:               policyProcessID,
		PassphraseAuthenticationToken: otp.AuthenticationToken,
		EncryptedPassphrase:           encrypted,
	}

	var response Response
	req, err := a.newVerifiedRequest(ctx, http.MethodPost, "stoic/validation", payload)
	if err != nil {
		return err
	}

	if err := a.do(req, &response, nil); err != nil {
		return err
	}

	if response.Otp.AcceptanceStatus != "ACCEPTED" {
		return fmt.Errorf("otp not accepted: %s", response.Otp.AcceptanceStatus)
	}

	if response.Otp.ProfileStatus != "UNLOCKED" {
		return fmt.Errorf("otp accepted but profile not unlocked: %s", response.Otp.ProfileStatus)
	}

	return nil
}

type ChallengeVerificationOtp struct {
	AuthenticationToken string `json:"authenticationToken"`
	EncryptionKey       string `json:"encryptionKey"`
}

type ChallengeVerificationResponse struct {
	Authenticator   string                   `json:"authenticator"`
	Otp             ChallengeVerificationOtp `json:"otp"`
	PolicyProcessID string                   `json:"policyProcessId"`
}

func (a *Web) ChallengeVerification(ctx context.Context, policyProcessID string, contactPoint ChallengeContactPoint) (ChallengeVerificationResponse, error) {
	type SelectedContactPoint struct {
		ID             string `json:"id"`
		DeliveryMedium string `json:"deliveryMedium"`
		MaskedValue    string `json:"maskedValue"`
	}

	type Payload struct {
		BusinessEvent        string               `json:"businessEvent"`
		ChallengeMethod      string               `json:"challengeMethod"`
		PolicyProcessID      string               `json:"policyProcessId"`
		SelectedContactPoint SelectedContactPoint `json:"selectedContactPoint"`
	}

	payload := Payload{
		BusinessEvent:   verificationBusinessEvent,
		ChallengeMethod: "OTP",
		PolicyProcessID: policyProcessID,
		SelectedContactPoint: SelectedContactPoint{
			ID:             contactPoint.ContactPointID,
			DeliveryMedium: "SMS",
			MaskedValue:    contactPoint.ContactPointMasked,
		},
	}

	req, err := a.newVerifiedRequest(ctx, http.MethodPost, "stoic/verification", payload)
	if err != nil {
		return ChallengeVerificationResponse{}, err
	}

	var response ChallengeVerificationResponse
	if err := a.do(req, &response, nil); err != nil {
		return ChallengeVerificationResponse{}, err
	}

	return response, nil

}

type ChallengeContactPointDeliveryMediums struct {
	IsSms bool `json:"isSms"`
}

type ChallengeContactPoint struct {
	ContactPointDeliveryMediums ChallengeContactPointDeliveryMediums `json:"contactPointDeliveryMediums"`
	ContactPointID              string                               `json:"contactPointId"`
	ContactPointLabel           string                               `json:"contactPointLabel"`
	ContactPointMasked          string                               `json:"contactPointMasked"`
}

type ChallengeMethodPayload struct {
	ContactPoints []ChallengeContactPoint `json:"contactPoints"`
}

type ChallengeMethod struct {
	Authenticator           string                 `json:"authenticator"`
	AvailableMethodsPayload ChallengeMethodPayload `json:"availableMethodsPayload"`
	IsLegacy                bool                   `json:"isLegacy"`
}

type ChallengeAssessment struct {
	AvailableMethods []ChallengeMethod `json:"availableMethods"`
	PolicyProcessID  string            `json:"policyProcessId"`
}

func (a *Web) ChallengeAssessment(ctx context.Context, card extension.PaymentCard) (ChallengeAssessment, error) {
	type WebRiskAssessment struct {
		DeviceFingerPrint string `json:"deviceFingerPrint"`
	}

	type RiskAssessment struct {
		Web WebRiskAssessment `json:"web"`
	}

	type Payload struct {
		JourneyID             string         `json:"journeyID"`
		RiskAssessment        RiskAssessment `json:"riskAssessment"`
		IntegrationParameters string         `json:"integrationParameters"`
		BusinessEvent         string         `json:"businessEvent"`
		GotoURL               string         `json:"gotoUrl"`
	}

	type contextReference struct {
		ReferenceID     string `json:"referenceId"`
		ReferenceIDType string `json:"referenceIdType"`
	}

	contextReferences, err := json.Marshal([]contextReference{
		{
			ReferenceID:     card.CardReferenceID,
			ReferenceIDType: "accountReferenceId",
		},
	})
	if err != nil {
		return ChallengeAssessment{}, err
	}

	cardRef := url.QueryEscape(card.CardReferenceID)
	gotoUrl := fmt.Sprintf("https://myaccounts.capitalone.com/VirtualCards/Manager/createVirtualCard?cardRef=%s&analyticsTag=from_more_account_services&pageIndex=0&pageSize=50&account=%s", cardRef, cardRef)
	encodedGotoUrl := base64.StdEncoding.EncodeToString([]byte(gotoUrl))

	integrationParameters := fmt.Sprintf("?businessEvent=%s&gotoUrl=%s&contextReferences=%s", verificationBusinessEvent, encodedGotoUrl, base64.StdEncoding.EncodeToString(contextReferences))

	encodedGotoUrl = strings.ReplaceAll(encodedGotoUrl, "/", "_")
	encodedGotoUrl = strings.ReplaceAll(encodedGotoUrl, "+", "-")
	encodedGotoUrl = strings.ReplaceAll(encodedGotoUrl, "=", "")

	fingerprint, err := a.api.GenerateFingerprintString("https://verified.capitalone.com/step-up/" + integrationParameters)
	if err != nil {
		return ChallengeAssessment{}, err
	}

	payload := Payload{
		JourneyID:             a.getTid(),
		RiskAssessment:        RiskAssessment{Web: WebRiskAssessment{DeviceFingerPrint: fingerprint}},
		IntegrationParameters: integrationParameters,
		BusinessEvent:         verificationBusinessEvent,
		GotoURL:               encodedGotoUrl,
	}

	var challenge ChallengeAssessment
	req, err := a.newVerifiedRequest(ctx, http.MethodPost, "stoic/challengeassessment", payload)
	if err != nil {
		return challenge, err
	}

	if err := a.do(req, &challenge, nil); err != nil {
		return challenge, err
	}

	return challenge, nil
}

func generateTid() string {
	return "IDEX-SIC-" + uuid.NewString()
}

// thread unsafe
func (a *Web) getTid() string {
	existingTid := a.api.GetCookie("c1_ubatid")
	if existingTid != "" {
		return existingTid
	}

	tid := generateTid()
	a.api.SetCookies([]*http.Cookie{
		{
			Name:  "c1_ubatid",
			Value: tid,
		},
	})
	return tid
}
