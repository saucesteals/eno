package extension

import (
	"context"
	"fmt"
	"net/http"
)

type OPTSmsContactDetails struct {
	ContactPointType string `json:"contactPointType"`
	PrimaryIndicator bool   `json:"primaryIndicator"`
	ContactPoint     string `json:"contactPoint"`
	ID               int    `json:"id"`
	Primary          bool   `json:"primary"`
}

type OTPOptions struct {
	SmsContactDetails []OPTSmsContactDetails `json:"smsContactDetails"`
}

func (a *Extension) OTPGenerate(ctx context.Context) (OTPOptions, error) {
	type Payload struct {
		ProfileReferenceID  string `json:"profileReferenceID"`
		HeaderFRCookie      string `json:"headerFRCookie"`
		ClientIPAddress     string `json:"clientIPAddress"`
		ClientCorrelationID string `json:"clientCorrelationID"`
		IsExpressLogin      bool   `json:"isExpressLogin"`
	}

	payload := Payload{
		ProfileReferenceID:  a.session.GetProfileReferenceId(),
		HeaderFRCookie:      a.session.GetRsaToken(),
		ClientIPAddress:     a.session.GetIpAddress(),
		ClientCorrelationID: a.clientReferenceId,
		IsExpressLogin:      false,
	}

	var response OTPOptions
	req, err := a.newWibRequest(ctx, http.MethodPost, "wib/challenge/otp/generate", payload)
	if err != nil {
		return response, err
	}

	if err := a.do(req, &response); err != nil {
		return response, err
	}

	return response, nil
}

type OTPAuthentication struct {
	Pin                    string `json:"pin"`
	PinAuthenticationToken string `json:"pinAuthenticationToken"`
	ContactPoint           string `json:"contactPoint"`
	ContactPointType       string `json:"contactPointType"`
}

func (a *Extension) OTPSend(ctx context.Context, contactPoint string) (OTPAuthentication, error) {
	type Payload struct {
		SelectedContactPoint string `json:"selectedContactPoint"`
		ProfileReferenceID   string `json:"profileReferenceID"`
		TransactionInfo      string `json:"transactionInfo"`
		HeaderFRCookie       string `json:"headerFRCookie"`
		IsExpressLogin       bool   `json:"isExpressLogin"`
	}

	payload := Payload{
		ProfileReferenceID:   a.session.GetProfileReferenceId(),
		HeaderFRCookie:       a.session.GetRsaToken(),
		SelectedContactPoint: contactPoint,
		TransactionInfo:      a.session.GetTransactionInfo(),
		IsExpressLogin:       false,
	}

	var response OTPAuthentication
	req, err := a.newWibRequest(ctx, http.MethodPost, "wib/challenge/otp/send", payload)
	if err != nil {
		return response, err
	}

	if err := a.do(req, &response); err != nil {
		return response, err
	}

	return response, nil
}

type OTPResult struct {
	ProfileStatus                 string `json:"profileStatus"`
	AcceptanceStatus              string `json:"acceptanceStatus"`
	ValidationStatus              string `json:"validationStatus"`
	PinAuthenticationFailureCount int    `json:"pinAuthenticationFailureCount"`
	UpgradedForgeRockCookie       string `json:"upgradedForgeRockCookie"`
}

func (a *Extension) OTPValidate(ctx context.Context, pin string, pinAuthenticationToken string) (OTPResult, error) {
	type Payload struct {
		Pin                    string `json:"pin"`
		ProfileReferenceID     string `json:"profileReferenceID"`
		HeaderFRCookie         string `json:"headerFRCookie"`
		DeviceExtensionid      string `json:"deviceExtensionid"`
		PinAuthenticationToken string `json:"pinAuthenticationToken"`
		IsExpressLogin         bool   `json:"isExpressLogin"`
	}

	payload := Payload{
		ProfileReferenceID:     a.session.GetProfileReferenceId(),
		HeaderFRCookie:         a.session.GetRsaToken(),
		Pin:                    pin,
		PinAuthenticationToken: pinAuthenticationToken,
		DeviceExtensionid:      a.device.ExtensionId,
		IsExpressLogin:         false,
	}

	var response OTPResult
	req, err := a.newWibRequest(ctx, http.MethodPost, "wib/challenge/otp/validate", payload)
	if err != nil {
		return response, err
	}

	if err := a.do(req, &response); err != nil {
		return response, err
	}

	if response.AcceptanceStatus != "ACCEPTED" {
		return response, fmt.Errorf("acceptance status: %s", response.AcceptanceStatus)
	}

	if response.ProfileStatus != "UNLOCKED" {
		return response, fmt.Errorf("profile status: %s", response.ProfileStatus)
	}

	if response.ValidationStatus != "SUCCESS" {
		return response, fmt.Errorf("validation status: %s", response.ValidationStatus)
	}

	if response.UpgradedForgeRockCookie != "" {
		a.session.SetRsaToken(response.UpgradedForgeRockCookie)
	}

	return response, nil
}
