package extension

import (
	"context"
	"net/http"
	"sync"
)

type LoginStatus string

var (
	LoginStatusSuccess   LoginStatus = "SUCCESS"
	LoginStatusChallenge LoginStatus = "CHALLENGE"
)

type Session struct {
	CustomerName          string      `json:"customerName"`
	ProfileReferenceID    string      `json:"profileReferenceID"`
	LoginStatus           LoginStatus `json:"loginStatus"`
	HeaderForgeRockCookie string      `json:"headerForgeRockCookie"`
	TransactionInfo       string      `json:"transactionInfo"`
}

func (a *Extension) GetSession(ctx context.Context) (*Session, error) {
	type Payload struct {
		DeviceFingerPrint string `json:"deviceFingerPrint"`
		DeviceExtensionid string `json:"deviceExtensionid"`
		AuthTransactionID string `json:"authTransactionId"`
		IsExpressLogin    bool   `json:"isExpressLogin"`
	}

	payload := Payload{
		DeviceFingerPrint: a.fingerprint,
		AuthTransactionID: a.device.AuthTransactionId,
		DeviceExtensionid: a.device.ExtensionId,
		IsExpressLogin:    false,
	}

	req, err := a.newWibRequest(ctx, http.MethodPost, "wib/user/session", payload)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := a.do(req, &session); err != nil {
		return nil, err
	}

	a.session.SetProfileReferenceId(session.ProfileReferenceID)
	a.session.SetRsaToken(session.HeaderForgeRockCookie)
	a.session.SetTransactionInfo(session.TransactionInfo)

	return &session, nil
}

type sessionDetails struct {
	mu                 sync.Mutex
	rsaToken           string
	accessToken        string
	profileReferenceId string
	ipAddress          string
	transactionInfo    string
}

func (a *sessionDetails) GetTransactionInfo() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.transactionInfo
}

func (a *sessionDetails) GetAccessToken() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.accessToken
}

func (a *sessionDetails) GetProfileReferenceId() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.profileReferenceId
}

func (a *sessionDetails) GetRsaToken() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.rsaToken
}

func (a *sessionDetails) GetIpAddress() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ipAddress
}

func (a *sessionDetails) SetTransactionInfo(info string) {
	a.mu.Lock()
	a.transactionInfo = info
	a.mu.Unlock()
}

func (a *sessionDetails) SetAccessToken(token string) {
	a.mu.Lock()
	a.accessToken = token
	a.mu.Unlock()
}

func (a *sessionDetails) SetProfileReferenceId(id string) {
	a.mu.Lock()
	a.profileReferenceId = id
	a.mu.Unlock()
}

func (a *sessionDetails) SetRsaToken(token string) {
	a.mu.Lock()
	a.rsaToken = token
	a.mu.Unlock()
}

func (a *sessionDetails) SetIpAddress(ip string) {
	a.mu.Lock()
	a.ipAddress = ip
	a.mu.Unlock()
}
