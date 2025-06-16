package extension

import (
	"context"
	"net/http"
)

type ExpressEnrollment struct {
	ExpressCheckoutToken string `json:"expressCheckoutToken"`
}

func (a *Extension) ExpressEnroll(ctx context.Context) (ExpressEnrollment, error) {
	type Payload struct {
		AuthenticationMethod   string `json:"authenticationMethod"`
		RsaToken               string `json:"rsaToken"`
		AuthenticationDeviceID string `json:"authenticationDeviceId"`
	}

	var response ExpressEnrollment
	deviceId, err := a.Encrypt(ctx, a.device.ExtensionId)
	if err != nil {
		return response, err
	}

	payload := Payload{
		AuthenticationMethod:   "Seamless-Web",
		RsaToken:               a.session.GetRsaToken(),
		AuthenticationDeviceID: deviceId,
	}

	req, err := a.newWibRequest(ctx, http.MethodPost, "wib/express-checkout/enroll", payload)
	if err != nil {
		return response, err
	}

	if err := a.do(req, &response); err != nil {
		return response, err
	}

	return response, nil
}

type ExpressLogin struct {
	ExpressCheckoutToken string `json:"expressCheckoutToken"`
}

func (a *Extension) ExpressLogin(ctx context.Context, expressToken string) (ExpressLogin, error) {
	type Payload struct {
		UserName               string `json:"userName"`
		DeviceFingerPrint      string `json:"deviceFingerPrint"`
		ExpressToken           string `json:"expressToken"`
		AuthenticationDeviceID string `json:"authenticationDeviceId"`
		DeviceExtensionid      string `json:"deviceExtensionid"`
	}

	var response ExpressLogin
	username, err := a.Encrypt(ctx, a.api.GetCredentials().Username)
	if err != nil {
		return response, err
	}

	deviceId, err := a.Encrypt(ctx, a.device.ExtensionId)
	if err != nil {
		return response, err
	}

	fingerprint, err := a.generateFingerprint()
	if err != nil {
		return response, err
	}

	payload := Payload{
		UserName:               username,
		DeviceFingerPrint:      fingerprint,
		ExpressToken:           expressToken,
		AuthenticationDeviceID: deviceId,
		DeviceExtensionid:      a.device.ExtensionId,
	}

	req, err := a.newWibRequest(ctx, http.MethodPost, "wib/express-login", payload)
	if err != nil {
		return response, err
	}

	if err := a.do(req, &response); err != nil {
		return response, err
	}

	return response, nil
}
