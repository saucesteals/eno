package extension

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	"github.com/saucesteals/eno/api"
)

var (
	chromeExtensionId      = "clmkdohmabikagpnhjmgacbclihgmdje"
	chromeExtensionVersion = "5.1.1"
)

func GetChromeExtensionVersion() string {
	return chromeExtensionVersion
}

func GetChromeExtensionId() string {
	return chromeExtensionId
}

func GetChromeExtensionURL() string {
	return "chrome-extension://" + GetChromeExtensionId()
}

type Extension struct {
	api *api.API

	device            Device
	fingerprint       string
	clientReferenceId string
	headers           http.Header
	session           *sessionDetails

	muKeys sync.Mutex
}

func New(api *api.API, device Device) (*Extension, error) {
	fingerprint := generateFingerprint(device)

	fp, err := json.Marshal(fingerprint)
	if err != nil {
		return nil, err
	}

	clientReferenceId := generateClientCorrelationId(device)

	return &Extension{
		api:               api,
		headers:           http.Header{},
		device:            device,
		fingerprint:       url.QueryEscape(base64.StdEncoding.EncodeToString(fp)),
		clientReferenceId: clientReferenceId,
		session:           &sessionDetails{},
	}, nil
}
