package extension

import (
	"net/http"
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
	clientReferenceId string
	headers           http.Header
	session           *sessionDetails

	muKeys sync.Mutex
}

func New(api *api.API, device Device) (*Extension, error) {
	clientReferenceId := generateClientCorrelationId(device, api.GetUserAgent())

	return &Extension{
		api:               api,
		headers:           http.Header{},
		device:            device,
		clientReferenceId: clientReferenceId,
		session:           &sessionDetails{},
	}, nil
}
