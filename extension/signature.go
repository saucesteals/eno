package extension

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

type Signature struct {
	Mac               string `json:"mac"`
	SessionIdentifier string `json:"sessionIdentifier"`
	ExtensionID       string `json:"extensionId"`
	Nonce             string `json:"nonce"`
	Endpoint          string `json:"endpoint"`
}

func (a *Extension) sign(keys *EncryptionKeys, endpoint string) (Signature, error) {
	nonceBytes := make([]byte, 16)
	_, _ = rand.Read(nonceBytes)
	nonce := hex.EncodeToString(nonceBytes)

	payload := strings.Join([]string{
		endpoint,
		a.clientReferenceId,
		a.device.ExtensionId,
		nonce,
	}, "")

	mac := keys.Sign([]byte(payload))
	macString := base64.StdEncoding.EncodeToString(mac)

	return Signature{
		Mac:               macString,
		SessionIdentifier: a.clientReferenceId,
		ExtensionID:       a.device.ExtensionId,
		Nonce:             nonce,
		Endpoint:          endpoint,
	}, nil
}
